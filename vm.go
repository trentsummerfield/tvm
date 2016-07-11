package java

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var debug = true

type VM struct {
	classes       []*Class
	dirs          []string
	activeMethod  *Method
	nativeMethods map[string](func(*VM, *Frame, io.Writer))
	frame         *Frame
	stdout        io.Writer
}

type Frame struct {
	stack
	PreviousFrame *Frame
	Class         *Class
	Method        *Method
	PC            *ProgramCounter
	Variables     []javaValue
	Root          bool
}

func NewVM() (vm VM) {
	vm.nativeMethods = map[string](func(*VM, *Frame, io.Writer)){
		"print":                   nativePrintString,
		"printInt":                nativePrintInteger,
		"printLong":               nativePrintLong,
		"printFloat":              nativePrintFloat,
		"printChar":               nativePrintChar,
		"arraycopy":               nativeArrayCopy,
		"desiredAssertionStatus0": nativeDesiredAssertionStatus,
		"fillInStackTrace":        nativeFillInStackTrace,
		"registerNatives":         nativeRegisterNatives,
		"getClass":                nativeGetClass,
	}
	return vm
}

func (vm *VM) LoadedClasses() []*Class {
	return vm.classes
}

func (vm *VM) ActiveMethod() *Method {
	return vm.activeMethod
}

func (vm *VM) ActiveFrame() *Frame {
	return vm.frame
}

func (vm *VM) AddDirectory(dir string) {
	vm.dirs = append(vm.dirs, dir)
}

func (vm *VM) LoadClass(path string) error {
	file, _ := os.Open(path)
	defer file.Close()
	class, err := ParseClass(file)
	if err != nil {
		return err
	}
	vm.classes = append(vm.classes, &class)
	return nil
}

func (vm *VM) Run() {
	vm.stdout = os.Stdout
	var frame Frame
	//TODO: push the actual command line arguments onto the stack for the main method.
	frame.push(nil)
	frame.Root = true
	//TODO: this is obviously completely wrong
	for _, c := range vm.classes {
		if c.hasMethodCalled("main") {
			vm.activeMethod = c.resolveMethod("main", "([Ljava/lang/String;)V")
		}
	}
	vm.execute(vm.activeMethod.class.Name(), vm.activeMethod.Name(), vm.activeMethod.RawSigniture, &frame, false, true)
}

func (vm *VM) Start() {
	vm.stdout = new(bytes.Buffer)
	var frame Frame
	//TODO: push the actual command line arguments onto the stack for the main method.
	frame.push(nil)
	//TODO: this is obviously completely wrong
	for _, c := range vm.classes {
		if c.hasMethodCalled("main") {
			vm.activeMethod = c.resolveMethod("main", "([Ljava/lang/String;)V")
		}
	}
	vm.execute(vm.activeMethod.class.Name(), vm.activeMethod.Name(), vm.activeMethod.RawSigniture, &frame, false, false)
}

func nativeStringToJavaString(vm *VM, str string) javaObject {
	c := vm.resolveClass("java/lang/String")
	ref := newInstance(c)
	arrLen := len(str)
	arr := make([]javaValue, arrLen)
	for i, _ := range arr {
		arr[i] = javaByte(str[i])
	}
	ref.fields["value"] = javaArray{contents: arr}
	ref.fields["count"] = javaInt(arrLen)
	return ref
}

func javaStringToNativeString(str javaObject) string {
	f := str.getField("value", "[c").(javaArray)
	bytes := make([]byte, len(f.contents))
	for i, b := range f.contents {
		bytes[i] = byte(b.(javaByte))
	}
	return string(bytes)
}

func nativePrintString(_ *VM, f *Frame, w io.Writer) {
	str := f.Variables[0]
	switch str := str.(type) {
	default:
		log.Fatalf("unexpected type %T", str)
	case javaObject:
		fmt.Fprint(w, javaStringToNativeString(str))
	}
	return
}

func nativePrintInteger(_ *VM, f *Frame, w io.Writer) {
	i := f.Variables[0].(javaInt).unbox()
	fmt.Fprintln(w, i)
	return
}

func nativePrintLong(_ *VM, f *Frame, w io.Writer) {
	i := f.Variables[0].(javaLong).unbox()
	fmt.Fprintln(w, i)
	return
}

func nativePrintFloat(_ *VM, f *Frame, w io.Writer) {
	float := f.Variables[0].(javaFloat).unbox()
	fmt.Fprintln(w, float)
	return
}

func nativePrintChar(_ *VM, f *Frame, w io.Writer) {
	i := f.Variables[0].(javaInt).unbox()
	fmt.Fprintf(w, "%c", i)
	return
}

func nativeArrayCopy(_ *VM, f *Frame, w io.Writer) {
	// (Object src,  int  srcPos, Object dest, int destPos, int length)
	src := f.Variables[0].(javaArray).contents
	i := f.Variables[1].(javaInt).unbox()
	dst := f.Variables[2].(javaArray).contents
	j := f.Variables[3].(javaInt).unbox()
	k := f.Variables[4].(javaInt).unbox()

	for l := int32(0); l < k; l++ {
		dst[j+l] = src[i+l]
	}

	return
}

func nativeDesiredAssertionStatus(_ *VM, f *Frame, w io.Writer) {
	f.PreviousFrame.pushInt32(0)
	return
}

func nativeFillInStackTrace(_ *VM, f *Frame, w io.Writer) {
	// (Throwable t, int unused)
	t := f.Variables[0].(javaObject)
	t.setField("stackTrace", javaArray{null: true})
	f.PreviousFrame.push(t)
	return
}

func nativeRegisterNatives(vm *VM, f *Frame, w io.Writer) {
}

func (vm *VM) setupSystemClass() {
	fd := vm.resolveClass("java/io/FileDescriptor")
	fileStream := vm.construct("java/io/FileInputStream", fd.getField("in").value)
	bufferedInputStream := vm.construct("java/io/BufferedInputStream", fileStream)
	system := vm.resolveClass("java/lang/System")
	system.getField("in").value = bufferedInputStream
}

func (vm *VM) construct(className string, arguments ...javaValue) javaObject {
	var frame Frame
	frame.Root = true
	class := vm.resolveClass(className)
	o := newInstance(class)
	frame.push(o)
	var descriptors []string
	for _, arg := range arguments {
		frame.push(arg)
		switch arg := arg.(type) {
		case javaObject:
			descriptors = append(descriptors, "L"+arg.class().Name())
		default:
			log.Fatalf("Fuck i don't know how to convert a %T to a descriptor\n", arg)
		}
	}
	descriptor := "(" + strings.Join(descriptors, ",") + ")L" + className + ";"
	vm.execute(className, "<init>", descriptor, &frame, false, true)
	return o
}

func nativeGetClass(vm *VM, f *Frame, w io.Writer) {
	o := f.Variables[0].(javaObject)
	classObject := newInstance(vm.resolveClass("java/lang/Class"))
	classObject.setField("name", nativeStringToJavaString(vm, o.class().Name()))
	f.PreviousFrame.push(classObject)
	return
}

func (vm *VM) resolveClass(name string) *Class {
	if strings.HasPrefix(name, "[L") {
		l := len(name)
		name = name[2 : l-1]
	}

	class := vm.getClass(name)
	if class != nil {
		return class
	}
	for _, d := range vm.dirs {
		err := vm.LoadClass(filepath.Join(d, name) + ".class")
		if err != nil {
			continue
		}
	}
	class = vm.getClass(name)
	if class != nil {
		return class
	}
	//TODO: raise the appropriate java exception
	log.Panicf("Could not resolve class %s\n", name)
	return nil
}

func (vm *VM) getClass(name string) *Class {
	for _, class := range vm.classes {
		// TODO: handle packages correctly
		if class.Name() == name {
			return class
		}
	}
	return nil
}

func collectArgs(method *Method, frame *Frame) []javaValue {
	numArgs := method.numArgs()
	if (method.accessFlags & Static) != 0 {
		numArgs--
	}
	args := make([]javaValue, numArgs+1)
	for i := numArgs; i >= 0; i-- {
		args[i] = frame.pop()
	}
	return args
}

func newFrame(previousFrame *Frame, method *Method, args []javaValue) Frame {
	var frame Frame
	var Variables int
	if method.Native() {
		Variables = method.numArgs()
		if !method.Static() {
			Variables += 1
		}
	} else {
		Variables = int(method.Code.maxLocals)
	}
	frame.Variables = make([]javaValue, Variables)
	offset := 0
	for i, v := range args {
		frame.Variables[i+offset] = v
		_, isLong := v.(javaLong)
		if isLong {
			offset++
		}
		_, isDouble := v.(javaDouble)
		if isDouble {
			offset++
		}
	}
	frame.Class = method.Class()
	frame.Method = method
	frame.PreviousFrame = previousFrame
	pc := newProgramCounter(method.Code.Instructions)
	frame.PC = &pc
	return frame
}

func buildFrame(vm *VM, className, methodName, descriptor string, previousFrame *Frame, virtual bool) *Frame {
	class := vm.resolveClass(className)
	method := class.resolveMethod(methodName, descriptor)
	if method == nil {
		if className == "java/lang/Object" {
			log.Fatalf("Unable to find method %s\n", methodName)
		}
		return buildFrame(vm, class.getSuperName(), methodName, descriptor, previousFrame, false)
	}
	args := collectArgs(method, previousFrame)
	if virtual {
		o := args[0].(javaObject)
		for _, v := range args {
			previousFrame.push(v)
		}
		return buildFrame(vm, o.class().Name(), methodName, descriptor, previousFrame, false)
	}

	frame := newFrame(previousFrame, method, args)
	return &frame
}

func (vm *VM) execute(className, methodName, descriptor string, previousFrame *Frame, virtual bool, run bool) {
	vm.frame = buildFrame(vm, className, methodName, descriptor, previousFrame, virtual)

	if run {
		for !vm.frame.Root {
			vm.Step()
		}
	}
}

func (vm *VM) Step() {
	if !vm.frame.Root {
		if (vm.frame.Method.accessFlags & Native) != 0 {
			methodName := vm.frame.Method.Name()
			native := vm.nativeMethods[methodName]
			if native == nil {
				log.Panicf("Unknown native method %s", methodName)
			}
			native(vm, vm.frame, vm.stdout)
			vm.frame = vm.frame.PreviousFrame
		} else {
			vm.frame = runByteCode(vm, vm.frame)
		}
	}
}

func runByteCode(vm *VM, frame *Frame) *Frame {
	op := frame.PC.next()
	switch op.name {
	case "nop":
	case "aconst_null":
		frame.push(javaObject{null: true})
	case "iconst_0":
		frame.pushInt32(0)
	case "iconst_1":
		frame.pushInt32(1)
	case "iconst_2":
		frame.pushInt32(2)
	case "iconst_3":
		frame.pushInt32(3)
	case "iconst_4":
		frame.pushInt32(4)
	case "iconst_5":
		frame.pushInt32(5)
	case "fconst_2":
		frame.pushFloat32(2.0)
	case "dconst_1":
		frame.pushFloat64(1.0)
	case "bipush":
		frame.pushInt32(int32(op.int8()))
	case "ldc":
		index := uint16(op.int8())
		switch constant := frame.Class.getConstantPoolItemAt(index).(type) {
		case intConstant:
			frame.pushInt32(constant.value)
		case floatConstant:
			frame.pushFloat32(constant.value)
		case stringConstant:
			s := frame.Class.getStringAt(int(index - 1))
			ref := nativeStringToJavaString(vm, s.contents)
			frame.push(ref)
		case classInfo:
			c := vm.resolveClass("java/lang/Class")
			ref := newInstance(c)
			frame.push(ref)
		default:
			log.Fatalf("Cannot load unknown constant %v", constant)
		}
	case "ldc2_w":
		index := op.int16()
		l := frame.Class.getLongAt(int(index - 1))
		frame.pushInt64(l.value)
	case "iload", "aload":
		index := op.int8()
		frame.push(frame.Variables[index])
	case "iload_0", "aload_0", "lload_0", "fload_0":
		frame.push(frame.Variables[0])
	case "iload_1", "aload_1", "lload_1", "fload_1":
		frame.push(frame.Variables[1])
	case "iload_2", "aload_2", "lload_2":
		frame.push(frame.Variables[2])
	case "iload_3", "aload_3", "lload_3":
		frame.push(frame.Variables[3])
	case "istore", "astore":
		index := op.int8()
		frame.Variables[index] = frame.pop()
	case "istore_1", "astore_1":
		frame.Variables[1] = frame.pop()
	case "istore_2", "astore_2":
		frame.Variables[2] = frame.pop()
	case "istore_3", "astore_3":
		frame.Variables[3] = frame.pop()
	case "castore":
		v := frame.popInt32()
		i := frame.popInt32()
		a := frame.popArray()
		a.contents[int(i)] = javaByte(v)
	case "caload":
		i := frame.popInt32()
		a := frame.popArray()
		c := byte(a.contents[int(i)].(javaByte))
		frame.pushInt32(int32(c))
	case "pop":
		frame.pop()
	case "dup":
		tmp := frame.pop()
		frame.push(tmp)
		frame.push(tmp)
	case "dup_x1":
		tmp1 := frame.pop()
		tmp2 := frame.pop()
		frame.push(tmp1)
		frame.push(tmp2)
		frame.push(tmp1)
	case "iadd":
		//TODO: make sure we do overflow correctly
		x := frame.popInt32()
		y := frame.popInt32()
		frame.pushInt32(x + y)
	case "isub":
		//TODO: make sure we do underflow correctly
		x := frame.popInt32()
		y := frame.popInt32()
		frame.pushInt32(y - x)
	case "imul":
		x := frame.popInt32()
		y := frame.popInt32()
		frame.pushInt32(x * y)
	case "idiv":
		x := frame.popInt32()
		y := frame.popInt32()
		frame.pushInt32(y / x)
	case "iinc":
		index := op.args[0]
		c := op.args[1]
		i := frame.Variables[index].(javaInt).unbox()
		frame.Variables[index] = javaInt(i + int32(c))
	case "ladd":
		//TODO: make sure we do overflow correctly
		x := frame.popInt64()
		y := frame.popInt64()
		frame.pushInt64(x + y)
	case "lsub":
		//TODO: make sure we do underflow correctly
		x := frame.popInt64()
		y := frame.popInt64()
		frame.pushInt64(y - x)
	case "lmul":
		x := frame.popInt64()
		y := frame.popInt64()
		frame.pushInt64(x * y)
	case "ldiv":
		x := frame.popInt64()
		y := frame.popInt64()
		frame.pushInt64(y / x)
	case "fadd":
		x := frame.popFloat32()
		y := frame.popFloat32()
		frame.pushFloat32(x + y)
	case "fsub":
		x := frame.popFloat32()
		y := frame.popFloat32()
		frame.pushFloat32(y - x)
	case "fmul":
		x := frame.popFloat32()
		y := frame.popFloat32()
		frame.pushFloat32(x * y)
	case "fdiv":
		x := frame.popFloat32()
		y := frame.popFloat32()
		frame.pushFloat32(y / x)
	case "ifeq":
		c := frame.popInt32()
		if c == 0 {
			frame.PC.jump(int(op.int16()))
		}
	case "ifne":
		c := frame.popInt32()
		if c != 0 {
			frame.PC.jump(int(op.int16()))
		}
	case "ifge":
		c := frame.popInt32()
		if c >= 0 {
			frame.PC.jump(int(op.int16()))
		}
	case "ifgt":
		c := frame.popInt32()
		if c > 0 {
			frame.PC.jump(int(op.int16()))
		}
	case "ifle":
		c := frame.popInt32()
		if c <= 0 {
			frame.PC.jump(int(op.int16()))
		}
	case "if_icmpne":
		v2 := frame.popInt32()
		v1 := frame.popInt32()
		if v1 != v2 {
			frame.PC.jump(int(op.int16()))
		}
	case "if_icmpge":
		v2 := frame.popInt32()
		v1 := frame.popInt32()
		if v1 >= v2 {
			frame.PC.jump(int(op.int16()))
		}
	case "if_icmpgt":
		v2 := frame.popInt32()
		v1 := frame.popInt32()
		if v1 > v2 {
			frame.PC.jump(int(op.int16()))
		}
	case "if_icmple":
		v2 := frame.popInt32()
		v1 := frame.popInt32()
		if v1 <= v2 {
			frame.PC.jump(int(op.int16()))
		}
	case "goto":
		frame.PC.jump(int(op.int16()))
	case "ireturn", "lreturn", "areturn", "freturn":
		frame.PreviousFrame.push(frame.pop())
		return frame.PreviousFrame
	case "return":
		return frame.PreviousFrame
	case "getstatic":
		fieldRef := frame.Class.getFieldRefAt(op.uint16())
		c := vm.resolveClass(fieldRef.className())
		vm.initClass(c, frame)
		f := c.getField(fieldRef.fieldName())
		frame.push(f.value)
	case "putstatic":
		fieldRef := frame.Class.getFieldRefAt(op.uint16())
		c := vm.resolveClass(fieldRef.className())
		f := c.getField(fieldRef.fieldName())
		f.value = frame.pop()
	case "getfield":
		fieldRef := frame.Class.getFieldRefAt(op.uint16())
		obj := frame.popObject()
		f := obj.getField(fieldRef.fieldName(), fieldRef.fieldDescriptor())
		frame.push(f)
	case "putfield":
		fieldRef := frame.Class.getFieldRefAt(op.uint16())
		f := frame.pop()
		obj := frame.popObject()
		obj.setField(fieldRef.fieldName(), f)
	case "invokevirtual":
		methodRef := frame.Class.getMethodRefAt(op.uint16())
		return buildFrame(vm, methodRef.className(), methodRef.methodName(), methodRef.methodType(), frame, true)
	case "invokespecial":
		methodRef := frame.Class.getMethodRefAt(op.uint16())
		return buildFrame(vm, methodRef.className(), methodRef.methodName(), methodRef.methodType(), frame, false)
	case "invokestatic":
		methodRef := frame.Class.getMethodRefAt(op.uint16())
		return buildFrame(vm, methodRef.className(), methodRef.methodName(), methodRef.methodType(), frame, false)
	case "invokeinterface":
		methodRef := frame.Class.getInterfaceMethodRefAt(op.uint16())
		return buildFrame(vm, methodRef.className(), methodRef.methodName(), methodRef.methodType(), frame, true)
	case "new":
		classInfo := frame.Class.getClassInfoAt(op.uint16())
		c := vm.resolveClass(classInfo.className())
		ref := newInstance(c)
		frame.push(ref)
	case "newarray":
		count := frame.popInt32()
		arr := make([]javaValue, count)
		for i, _ := range arr {
			arr[i] = javaByte(67)
		}
		frame.pushArray(javaArray{contents: arr})
	case "anewarray":
		classInfo := frame.Class.getClassInfoAt(op.uint16())
		count := frame.popInt32()
		arr := make([]javaValue, count)
		for i, _ := range arr {
			//TODO: set the class correctly
			arr[i] = javaObject{null: true}
		}
		frame.pushArray(javaArray{_class: vm.resolveClass(classInfo.className()), contents: arr})
	case "arraylength":
		a := frame.popArray()
		frame.pushInt32(int32(len(a.contents)))
	case "athrow":
		return handleException(vm, frame, frame.popObject())
	case "checkcast":
		o := frame.popReference()
		if o.isNull() {
			frame.pushReference(o)
		} else {
			classInfo := frame.Class.getClassInfoAt(op.uint16())
			c := classInfo.className()
			targetClass := vm.resolveClass(c)
			if vm.implements(o.class(), targetClass) {
				frame.pushReference(o)
			} else {
				//TODO: throw ClassCastException
				log.Fatal("Class Cast Exception should be thrown here")
			}
		}
	case "instanceof":
		classInfo := frame.Class.getClassInfoAt(op.uint16())
		c := classInfo.className()
		o := frame.popReference()
		targetClass := vm.resolveClass(c)
		if vm.implements(o.class(), targetClass) {
			frame.pushInt32(1)
		} else {
			frame.pushInt32(0)
		}
	case "monitorenter":
		//TODO: implement when we get to threading
	case "monitorexit":
		//TODO: implement when we get to threading
	case "ifnull":
		o := frame.popReference()
		if o.isNull() {
			frame.PC.jump(int(op.int16()))
		}
	case "ifnonnull":
		o := frame.popReference()
		if !o.isNull() {
			frame.PC.jump(int(op.int16()))
		}
	default:
		panic(fmt.Sprintf("Cannot execute instruction: %v", op))
	}
	return frame
}

func handleException(vm *VM, f *Frame, throwable javaObject) *Frame {
	if f.Root {
		f.push(throwable)
		vm.execute(throwable.class().Name(), "toString", "()Ljava/lang/String;", f, false, true)
		str := f.popObject()
		log.Fatalf("Unhandled exception %s\n", javaStringToNativeString(str))
	}
	index := f.PC.CurrentByteCodeIndex()
	match := false
	for _, handler := range f.Method.Code.ExceptionHandlers {
		if handler.CatchType == 0 || vm.implements(throwable.class(), vm.resolveClass(handler.Class)) {
			if index >= int(handler.Start) && index < int(handler.End) {
				f.PC.jumpTo(int(handler.Handler))
				match = true
				break
			}
		}
	}
	if match {
		f.push(throwable)
		return f
	} else {
		f = f.PreviousFrame
	}
	return handleException(vm, f, throwable)
}

func (vm *VM) implements(child *Class, parent *Class) bool {
	if child.Name() == parent.Name() {
		return true
	}

	for child.Name() != "java/lang/Object" {
		child = vm.resolveClass(child.getSuperName())
		if child.Name() == parent.Name() {
			return true
		}
	}

	return false
}

func newInstance(c *Class) javaObject {
	return javaObject{_class: c, fields: make(map[string]javaValue)}
}

func (vm *VM) initClass(c *Class, frame *Frame) {
	if c.initialised {
		return
	}
	c.initialised = true
	fOld := vm.frame
	var fNew Frame
	fNew.Root = true
	vm.execute(c.Name(), "<clinit>", "()V", &fNew, false, true)
	vm.frame = fOld
}

type stack struct {
	Items []javaValue
	size  uint
}

type javaValue interface {
	isJavaValue()
	String() string
}

func (s *stack) push(e javaValue) {
	s.Items = append(s.Items, e)
	s.size++
}

func (s *stack) pop() javaValue {
	if s.size == 0 {
		panic("Cannot pop from an empty stack")
	}
	e := s.Items[s.size-1]
	s.size--
	s.Items = s.Items[:s.size]
	return e
}

type javaInt int32

func (_ javaInt) isJavaValue() {}

func (v javaInt) String() string {
	return fmt.Sprintf("int(%v)", v.unbox())
}

func (s *stack) pushInt32(i int32) {
	s.push(javaInt(i))
}

func (s *stack) popInt32() int32 {
	return int32(s.pop().(javaInt))
}

func (i javaInt) unbox() int32 {
	return int32(i)
}

type javaLong int64

func (_ javaLong) isJavaValue() {}

func (l javaLong) String() string {
	return fmt.Sprintf("long(%v)", l.unbox())
}

func (s *stack) pushInt64(l int64) {
	s.push(javaLong(l))
}

func (s *stack) popInt64() int64 {
	return int64(s.pop().(javaLong))
}

func (i javaLong) unbox() int64 {
	return int64(i)
}

type javaFloat float32

func (_ javaFloat) isJavaValue() {}

func (f javaFloat) String() string {
	return fmt.Sprintf("float(%v)", f.unbox())
}

func (s *stack) pushFloat32(f float32) {
	s.push(javaFloat(f))
}

func (s *stack) popFloat32() float32 {
	return float32(s.pop().(javaFloat))
}

func (f javaFloat) unbox() float32 {
	return float32(f)
}

type javaDouble float64

func (_ javaDouble) isJavaValue() {}

func (v javaDouble) String() string {
	return fmt.Sprintf("double(%v)", v.unbox())
}

func (s *stack) pushFloat64(f float64) {
	s.push(javaDouble(f))
}

func (s *stack) popFloat64() float64 {
	return float64(s.pop().(javaDouble))
}

func (f javaDouble) unbox() float64 {
	return float64(f)
}

type javaByte byte

func (_ javaByte) isJavaValue() {}

func (v javaByte) String() string {
	return fmt.Sprintf("byte(%v)", v.unbox())
}

func (s *stack) pushByte(i byte) {
	s.push(javaByte(i))
}

func (s *stack) popByte() byte {
	return byte(s.pop().(javaByte))
}

func (b javaByte) unbox() byte {
	return byte(b)
}

type javaReference interface {
	isNull() bool
	class() *Class
}

func (s *stack) popReference() javaReference {
	return s.pop().(javaReference)
}

func (s *stack) pushReference(ref javaReference) {
	s.push(ref.(javaValue))
}

type javaArray struct {
	_class   *Class
	contents []javaValue
	null     bool
}

func (a javaArray) isNull() bool {
	return a.null
}

func (a javaArray) class() *Class {
	return a._class
}

func (_ javaArray) isJavaValue() {}

func (a javaArray) String() string {
	if a.isNull() {
		return fmt.Sprintf("Array(<null>)")
	}
	if a.class() != nil {
		return fmt.Sprintf("Array(%d,%s)", len(a.contents), a.class().Name())
	}
	return fmt.Sprintf("Array(%d,%T)", len(a.contents), a.contents)
}

func (s *stack) pushArray(a javaArray) {
	s.push(javaArray(a))
}

func (s *stack) popArray() javaArray {
	return s.pop().(javaArray)
}

type javaObject struct {
	null   bool
	_class *Class
	fields map[string]javaValue
}

func (o javaObject) isNull() bool {
	return o.null
}

func (o javaObject) class() *Class {
	return o._class
}

func (o *javaObject) getField(name, descriptor string) javaValue {
	_, ok := o.fields[name]
	if !ok {
		switch descriptor[0] {
		case 'I', 'Z':
			o.fields[name] = javaInt(0)
		case 'L':
			o.fields[name] = javaObject{null: true}
		case '[':
			o.fields[name] = javaArray{null: true}
		default:
			log.Fatalf("I don't know how to initialize field %v with type %v in %v\n", name, descriptor, o.class().Name())
		}
	}
	result := o.fields[name]
	return result
}

func (o *javaObject) setField(name string, f javaValue) {
	o.fields[name] = f
}

func (_ javaObject) isJavaValue() {}

func (o javaObject) String() string {
	if o.isNull() {
		return "Object(<null>)"
	}
	return fmt.Sprintf("Object(%s)", o.class().Name())
}

func (s *stack) popObject() javaObject {
	return s.pop().(javaObject)
}

func (f *Frame) DebugOut() {
	log.Printf("Stack %s::%s\n", f.Method.class.Name(), f.Method.Name())
	log.Print("=====")
	for i, o := range f.Items {
		log.Printf("%3d %v", i, o.String())
	}
}

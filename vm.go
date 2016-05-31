package java

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

type VM struct {
	classes       []*Class
	activeMethod  *Method
	nativeMethods map[string](func(*Frame, io.Writer))
	frame         *Frame
	stdout        io.Writer
}

type Frame struct {
	stack
	PreviousFrame *Frame
	Class         *Class
	Method        *Method
	PC            *ProgramCounter
	variables     []javaValue
	Root          bool
}

func NewVM() (vm VM) {
	vm.nativeMethods = map[string](func(*Frame, io.Writer)){
		"print":      nativePrintString,
		"printInt":   nativePrintInteger,
		"printLong":  nativePrintLong,
		"printFloat": nativePrintFloat,
		"printChar":  nativePrintChar,
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

func (vm *VM) LoadClass(path string) {
	file, _ := os.Open(path)
	class, err := ParseClass(file)
	if err != nil {
		log.Panicf("unable to load %s: %v", path, err)
	}
	vm.classes = append(vm.classes, &class)
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

func nativePrintString(f *Frame, w io.Writer) {
	str := f.variables[0]
	switch str := str.(type) {
	default:
		fmt.Printf("unexpected type %T", str)
	case javaObject:
		f := str.getField("value", "[c").(javaArray)
		bytes := make([]byte, len(f))
		for i, b := range f {
			bytes[i] = byte(b.(javaByte))
		}
		fmt.Fprint(w, string(bytes))
	}
	return
}

func nativePrintInteger(f *Frame, w io.Writer) {
	i := f.variables[0].(javaInt).unbox()
	fmt.Fprintln(w, i)
	return
}

func nativePrintLong(f *Frame, w io.Writer) {
	i := f.variables[0].(javaLong).unbox()
	fmt.Fprintln(w, i)
	return
}

func nativePrintFloat(f *Frame, w io.Writer) {
	float := f.variables[0].(javaFloat).unbox()
	fmt.Fprintln(w, float)
	return
}

func nativePrintChar(f *Frame, w io.Writer) {
	i := f.variables[0].(javaInt).unbox()
	fmt.Fprintf(w, "%c", i)
	return
}

func (vm *VM) resolveClass(name string) *Class {
	for _, class := range vm.classes {
		// TODO: handle packages correctly
		if class.Name() == name {
			return class
		}
	}
	//TODO: raise the appropriate java exception
	log.Panicf("Could not resolve class %s\n", name)
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
	if (method.accessFlags & Native) != 0 {
		frame.variables = make([]javaValue, method.numArgs())
	} else {
		frame.variables = make([]javaValue, method.Code.maxLocals)
	}
	offset := 0
	for i, v := range args {
		frame.variables[i+offset] = v
		_, isLong := v.(javaLong)
		if isLong {
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
		return buildFrame(vm, class.getSuperName(), methodName, descriptor, previousFrame, false)
	}
	args := collectArgs(method, previousFrame)
	if virtual {
		o := args[0].(javaObject)
		for _, v := range args {
			previousFrame.push(v)
		}
		return buildFrame(vm, o.className, methodName, descriptor, previousFrame, false)
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
			native(vm.frame, vm.stdout)
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
		frame.push(javaObject{isNull: true})
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
			c := vm.resolveClass("java/lang/String")
			ref := newInstance(c)
			arr := make([]javaValue, len(s.contents))
			for i, _ := range arr {
				arr[i] = javaByte(s.contents[i])
			}
			ref.fields["value"] = javaArray(arr)
			frame.push(ref)
		default:
			log.Fatalf("Cannot load unknown constant %v", constant)
		}
	case "ldc2_w":
		index := op.int16()
		l := frame.Class.getLongAt(int(index - 1))
		frame.pushInt64(l.value)
	case "iload":
		index := op.int8()
		frame.push(frame.variables[index])
	case "iload_0", "aload_0", "lload_0", "fload_0":
		frame.push(frame.variables[0])
	case "iload_1", "aload_1", "lload_1", "fload_1":
		frame.push(frame.variables[1])
	case "iload_2", "aload_2", "lload_2":
		frame.push(frame.variables[2])
	case "iload_3", "aload_3", "lload_3":
		frame.push(frame.variables[3])
	case "istore":
		index := op.int8()
		frame.variables[index] = frame.pop()
	case "istore_1", "astore_1":
		frame.variables[1] = frame.pop()
	case "istore_2", "astore_2":
		frame.variables[2] = frame.pop()
	case "istore_3", "astore_3":
		frame.variables[3] = frame.pop()
	case "castore":
		v := frame.popInt32()
		i := frame.popInt32()
		a := frame.popArray()
		a[int(i)] = javaByte(v)
	case "caload":
		i := frame.popInt32()
		a := frame.popArray()
		c := byte(a[int(i)].(javaByte))
		frame.pushInt32(int32(c))
	case "pop":
		frame.pop()
	case "dup":
		tmp := frame.pop()
		frame.push(tmp)
		frame.push(tmp)
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
		i := frame.variables[index].(javaInt).unbox()
		frame.variables[index] = javaInt(i + int32(c))
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
	case "ifle":
		c := frame.popInt32()
		if c <= 0 {
			frame.PC.jump(int(op.int16()))
		}
	case "if_icmpge":
		v2 := frame.popInt32()
		v1 := frame.popInt32()
		if v1 >= v2 {
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
		frame.pushArray(arr)
	case "arraylength":
		a := frame.popArray()
		frame.pushInt32(int32(len(a)))
	case "ifnull":
		o := frame.popObject()
		if o.isNull {
			frame.PC.jump(int(op.int16()))
		}
	case "ifnonnull":
		o := frame.popObject()
		if !o.isNull {
			frame.PC.jump(int(op.int16()))
		}
	default:
		panic(fmt.Sprintf("Cannot execute instruction: %v", op))
	}
	return frame
}

func newInstance(c *Class) javaObject {
	return javaObject{className: c.Name(), fields: make(map[string]javaValue)}
}

func (vm *VM) initClass(c *Class, frame *Frame) {
	if c.initialised {
		return
	}
	c.initialised = true
	vm.frame = buildFrame(vm, c.Name(), "<clinit>", "()V", frame, false)
	for vm.frame != frame {
		vm.Step()
	}
}

type stack struct {
	Items []javaValue
	size  uint
}

type javaValue interface {
	isJavaValue()
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

func (s *stack) pushFloat32(f float32) {
	s.push(javaFloat(f))
}

func (s *stack) popFloat32() float32 {
	return float32(s.pop().(javaFloat))
}

func (f javaFloat) unbox() float32 {
	return float32(f)
}

type javaByte byte

func (_ javaByte) isJavaValue() {}

func (s *stack) pushByte(i byte) {
	s.push(javaByte(i))
}

func (s *stack) popByte() byte {
	return byte(s.pop().(javaByte))
}

type javaArray []javaValue

func (_ javaArray) isJavaValue() {}

func (s *stack) pushArray(a javaArray) {
	s.push(javaArray(a))
}

func (s *stack) popArray() javaArray {
	return []javaValue(s.pop().(javaArray))
}

type javaObject struct {
	isNull    bool
	className string
	fields    map[string]javaValue
}

func (o *javaObject) getField(name, descriptor string) javaValue {
	_, ok := o.fields[name]
	if !ok {
		switch descriptor {
		case "I", "Z":
			o.fields[name] = javaInt(0)
		default:
			log.Fatalf("I don't know how to initialize field %v with type %v\n", name, descriptor)
		}
	}
	return o.fields[name]
}

func (o *javaObject) setField(name string, f javaValue) {
	o.fields[name] = f
}

func (_ javaObject) isJavaValue() {}

func (s *stack) popObject() javaObject {
	return s.pop().(javaObject)
}

package java

import (
	"fmt"
	"log"
	"os"
)

type VM struct {
	classes       []class
	nativeMethods map[string](func(*frame))
}

type frame struct {
	stack
	variables []javaValue
}

func NewVM() (vm VM) {
	vm.nativeMethods = map[string](func(*frame)){
		"print":     nativePrintString,
		"printInt":  nativePrintInteger,
		"printChar": nativePrintChar,
	}
	return vm
}

func (vm *VM) LoadClass(path string) {
	file, _ := os.Open(path)
	class, err := parseClass(file)
	if err != nil {
		log.Panicf("unable to load %s: %v", path, err)
	}
	vm.classes = append(vm.classes, class)
}

func (vm *VM) Run() {
	var frame frame
	//TODO: push the actual command line arguments onto the stack for the main method.
	frame.push(nil)
	//TODO: this is obviously completely wrong
	var index int
	for i, c := range vm.classes {
		if c.hasMethodCalled("main") {
			index = i
		}
	}
	vm.execute(vm.classes[index].getName(), "main", &frame, false)
}

func nativePrintString(f *frame) {
	str := f.variables[0]
	switch str := str.(type) {
	default:
		fmt.Printf("unexpected type %T", str)
	case javaObject:
		f := str.getField("data").(javaArray)
		bytes := make([]byte, len(f))
		for i, b := range f {
			bytes[i] = byte(b.(javaByte))
		}
		fmt.Print(string(bytes))
	}
	return
}

func nativePrintInteger(f *frame) {
	i := f.variables[0].(javaInt).unbox()
	fmt.Println(i)
	return
}

func nativePrintChar(f *frame) {
	i := f.variables[0].(javaInt).unbox()
	fmt.Printf("%c", i)
	return
}

func (vm *VM) resolveClass(name string) class {
	for _, class := range vm.classes {
		// TODO: handle packages correctly
		if class.getName() == name {
			return class
		}
	}
	//TODO: raise the appropriate java exception
	log.Panicf("Could not resolve class %s\n", name)
	return class{}
}

func collectArgs(method method, frame *frame) []javaValue {
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

func newFrame(method method, args []javaValue) frame {
	var frame frame
	if (method.accessFlags & Native) != 0 {
		frame.variables = make([]javaValue, method.numArgs())
	} else {
		frame.variables = make([]javaValue, method.code.maxLocals)
	}
	for i, v := range args {
		frame.variables[i] = v
	}
	return frame
}

func (vm *VM) execute(className string, methodName string, previousFrame *frame, virtual bool) {
	class := vm.resolveClass(className)
	if !class.hasMethodCalled(methodName) {
		vm.execute(class.getSuperName(), methodName, previousFrame, false)
		return
	}
	method := class.getMethod(methodName)
	args := collectArgs(method, previousFrame)
	if virtual {
		o := args[0].(javaObject)
		for _, v := range args {
			previousFrame.push(v)
		}
		vm.execute(o.className, methodName, previousFrame, false)
		return
	}
	frame := newFrame(method, args)

	if (method.accessFlags & Native) != 0 {
		native := vm.nativeMethods[methodName]
		if native == nil {
			log.Panicf("Unknown native method %s", methodName)
		}
		native(&frame)
		return
	}

	pc := newProgramCounter(method.code.code)
	for {
		op := pc.next()
		switch op.name {
		case "nop":
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
		case "bipush":
			frame.pushInt32(int32(op.int8()))
		case "ldc":
			s := class.getStringAt(int(op.int8() - 1))
			c := vm.resolveClass("java/lang/String")
			ref := newInstance(&c)
			arr := make([]javaValue, len(s.contents))
			for i, _ := range arr {
				arr[i] = javaByte(s.contents[i])
			}
			ref.fields["data"] = javaArray(arr)
			frame.push(ref)
		case "iload":
			index := op.int8()
			frame.push(frame.variables[index])
		case "iload_0", "aload_0":
			frame.push(frame.variables[0])
		case "iload_1", "aload_1":
			frame.push(frame.variables[1])
		case "iload_2", "aload_2":
			frame.push(frame.variables[2])
		case "iload_3", "aload_3":
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
		case "ifne":
			c := frame.popInt32()
			if c != 0 {
				pc.jump(int(op.int16()))
			}
		case "if_icmpge":
			v2 := frame.popInt32()
			v1 := frame.popInt32()
			if v1 >= v2 {
				pc.jump(int(op.int16()))
			}
		case "if_icmple":
			v2 := frame.popInt32()
			v1 := frame.popInt32()
			if v1 <= v2 {
				pc.jump(int(op.int16()))
			}
		case "goto":
			pc.jump(int(op.int16()))
		case "ireturn", "areturn":
			previousFrame.push(frame.pop())
			return
		case "return":
			return
		case "getstatic":
			fieldRef := class.getFieldRefAt(op.uint16())
			c := vm.resolveClass(fieldRef.className())
			vm.initClass(&c, &frame)
			f := c.getField(fieldRef.fieldName())
			frame.push(f.value)
		case "putstatic":
			fieldRef := class.getFieldRefAt(op.uint16())
			c := vm.resolveClass(fieldRef.className())
			f := c.getField(fieldRef.fieldName())
			f.value = frame.pop()
		case "getfield":
			fieldRef := class.getFieldRefAt(op.uint16())
			obj := frame.popObject()
			f := obj.getField(fieldRef.fieldName())
			frame.push(f)
		case "putfield":
			fieldRef := class.getFieldRefAt(op.uint16())
			f := frame.pop()
			obj := frame.popObject()
			obj.setField(fieldRef.fieldName(), f)
		case "invokevirtual":
			methodRef := class.getMethodRefAt(op.uint16())
			vm.execute(methodRef.className(), methodRef.methodName(), &frame, true)
		case "invokespecial":
			methodRef := class.getMethodRefAt(op.uint16())
			vm.execute(methodRef.className(), methodRef.methodName(), &frame, false)
		case "invokestatic":
			methodRef := class.getMethodRefAt(op.uint16())
			vm.execute(methodRef.className(), methodRef.methodName(), &frame, false)
		case "new":
			classInfo := class.getClassInfoAt(op.uint16())
			c := vm.resolveClass(classInfo.className())
			ref := newInstance(&c)
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
		default:
			panic(fmt.Sprintf("Cannot execute instruction: %v", op))
		}
	}
}

func newInstance(c *class) javaObject {
	return javaObject{c.getName(), make(map[string]javaValue)}
}

func (vm *VM) initClass(c *class, frame *frame) {
	if c.initialised {
		return
	}
	vm.execute(c.getName(), "<clinit>", frame, false)
	c.initialised = true
}

type stack struct {
	items []javaValue
	size  uint
}

type javaValue interface {
	isJavaValue()
}

func (s *stack) push(e javaValue) {
	s.items = append(s.items, e)
	s.size++
}

func (s *stack) pop() javaValue {
	if s.size == 0 {
		panic("Cannot pop from an empty stack")
	}
	e := s.items[s.size-1]
	s.size--
	s.items = s.items[:s.size]
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
	className string
	fields    map[string]javaValue
}

func (o *javaObject) getField(name string) javaValue {
	_, ok := o.fields[name]
	if !ok {
		o.fields[name] = javaInt(0)
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

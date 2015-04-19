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
		"print":    nativePrintString,
		"printInt": nativePrintInteger,
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
	frame.pushString(utf8String{""})
	//TODO: this is obviously completely wrong
	var index int
	for i, c := range vm.classes {
		if c.hasMethodCalled("main") {
			index = i
		}
	}
	vm.execute(vm.classes[index].getName(), "main", &frame)
}

func nativePrintString(f *frame) {
	str := f.variables[0].(utf8String)
	fmt.Print(str.contents)
	return
}

func nativePrintInteger(f *frame) {
	i := f.variables[0].(javaInt).unbox()
	fmt.Println(i)
	return
}

func (vm *VM) resolveClass(name string) class {
	for _, class := range vm.classes {
		if class.getName() == name {
			return class
		}
	}
	//TODO: raise the appropriate java exception
	log.Panicf("Could not resolve class %s\n", name)
	return class{}
}

func newFrame(method method, previousFrame *frame) frame {
	var frame frame
	numArgs := method.numArgs()
	if (method.accessFlags & Native) != 0 {
		frame.variables = make([]javaValue, method.numArgs())
	} else {
		frame.variables = make([]javaValue, method.code.maxLocals)
	}
	if (method.accessFlags & Static) != 0 {
		numArgs--
	}
	for i := numArgs; i >= 0; i-- {
		frame.variables[i] = previousFrame.pop()
	}
	return frame
}

func (vm *VM) execute(className string, methodName string, previousFrame *frame) {
	if className == "java/lang/Object" {
		return
	}
	class := vm.resolveClass(className)
	method := class.getMethod(methodName)
	frame := newFrame(method, previousFrame)

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
		case "iconst_1":
			frame.pushInt32(1)
		case "iconst_2":
			frame.pushInt32(2)
		case "iconst_5":
			frame.pushInt32(5)
		case "bipush":
			frame.pushInt32(int32(op.int8()))
		case "ldc":
			s := class.getStringAt(int(op.int8() - 1))
			frame.push(s)
		case "iload_0":
			frame.push(frame.variables[0])
		case "iload_1":
			frame.push(frame.variables[1])
		case "aload_0":
			frame.push(frame.variables[0].(javaObject))
		case "aload_1":
			frame.push(frame.variables[1].(javaObject))
		case "istore_1":
			frame.variables[1] = frame.pop()
		case "astore_1":
			frame.variables[1] = frame.pop().(javaObject)
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
		case "ifne":
			c := frame.popInt32()
			if c != 0 {
				pc.jump(int(op.int16()))
			}
		case "goto":
			pc.jump(int(op.int16()))
		case "ireturn":
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
			vm.execute(methodRef.className(), methodRef.methodName(), &frame)
		case "invokespecial":
			methodRef := class.getMethodRefAt(op.uint16())
			vm.execute(methodRef.className(), methodRef.methodName(), &frame)
		case "invokestatic":
			methodRef := class.getMethodRefAt(op.uint16())
			vm.execute(methodRef.className(), methodRef.methodName(), &frame)
		case "new":
			classInfo := class.getClassInfoAt(op.uint16())
			c := vm.resolveClass(classInfo.className())
			ref := newInstance(&c)
			frame.push(ref)
		default:
			panic(fmt.Sprintf("Cannot execute instruction: %v", op))
		}
	}
}

func newInstance(c *class) javaObject {
	return javaObject{make(map[string]javaValue)}
}

func (vm *VM) initClass(c *class, frame *frame) {
	if c.initialised {
		return
	}
	vm.execute(c.getName(), "<clinit>", frame)
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

func (_ utf8String) isJavaValue() {}

func (s *stack) pushString(str utf8String) {
	s.push(str)
}

func (s *stack) popString() utf8String {
	return s.pop().(utf8String)
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

type javaObject struct {
	fields map[string]javaValue
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

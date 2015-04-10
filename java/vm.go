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
	stack     stack
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
	frame.stack.pushString(utf8String{""})
	//TODO: this is obviously completely wrong
	vm.execute(vm.classes[0].getName(), "main", &frame)
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
	if (method.accessFlags & Native) != 0 {
		frame.variables = make([]javaValue, method.numArgs())
	} else {
		frame.variables = make([]javaValue, method.code.maxLocals)
	}
	for i := method.numArgs() - 1; i >= 0; i-- {
		frame.variables[i] = previousFrame.stack.pop()
	}
	return frame
}

type programCounter struct {
	code     *code
	position uint
}

func (code *code) newProgramCounter() programCounter {
	var pc programCounter
	pc.code = code
	return pc
}

func (pc *programCounter) nextByte() uint8 {
	b := pc.code.code[pc.position]
	pc.position++
	return b
}

func (vm *VM) execute(className string, methodName string, previousFrame *frame) {
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

	pc := method.code.newProgramCounter()
	for {
		instruction := pc.nextByte()
		switch instruction {
		case 0:
		case 5:
			frame.stack.pushInt32(2)
		case 8:
			frame.stack.pushInt32(5)
		case 16:
			frame.stack.pushInt32(int32(pc.nextByte()))
		case 18:
			s := class.getStringAt(int(pc.nextByte() - 1))
			frame.stack.push(s)
		case 26:
			frame.stack.push(frame.variables[0])
		case 27:
			frame.stack.push(frame.variables[1])
		case 60:
			frame.variables[1] = frame.stack.pop()
		case 96:
			//TODO: make sure we do overflow correctly
			x := frame.stack.popInt32()
			y := frame.stack.popInt32()
			frame.stack.pushInt32(x + y)
		case 100:
			//TODO: make sure we do underflow correctly
			x := frame.stack.popInt32()
			y := frame.stack.popInt32()
			frame.stack.pushInt32(y - x)
		case 104:
			x := frame.stack.popInt32()
			y := frame.stack.popInt32()
			frame.stack.pushInt32(x * y)
		case 108:
			x := frame.stack.popInt32()
			y := frame.stack.popInt32()
			frame.stack.pushInt32(y / x)
		case 172:
			previousFrame.stack.push(frame.stack.pop())
			return
		case 177:
			return
		case 184:
			var i uint16
			i |= uint16(pc.nextByte()) << 8
			i |= uint16(pc.nextByte())
			methodRef := class.getMethodRefAt(i)
			vm.execute(methodRef.className(), methodRef.methodName(), &frame)
		default:
			panic(fmt.Sprintf("Unknown instruction: %v", instruction))
		}
	}
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

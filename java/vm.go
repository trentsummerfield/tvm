package java

import (
	"fmt"
	"io/ioutil"
	"log"
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
	bytes, _ := ioutil.ReadFile(path)
	class, _ := parseClass(bytes)
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
	i := int32(f.variables[0].(javaInt))
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
			strRef := class.constantPoolItems[pc.nextByte()-1].(stringConstant)
			str := class.constantPoolItems[strRef.utf8Index-1].(utf8String)
			frame.stack.pushString(str)
		case 26:
			frame.stack.pushInt32(int32(frame.variables[0].(javaInt)))
		case 27:
			frame.stack.pushInt32(int32(frame.variables[1].(javaInt)))
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
			i = uint16(pc.nextByte()) << 8
			i |= uint16(pc.nextByte())
			methodRef := class.getMethodRefAt(i)
			vm.execute(methodRef.className(), methodRef.methodName(), &frame)
			break
		default:
			panic(fmt.Sprintf("Unknown instruction: %v", instruction))
		}
	}
}

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
	variables []stackItem
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
	vm.execute(vm.classes[0], "main", &frame)
}

func nativePrintString(f *frame) {
	str := f.variables[0].(utf8String)
	fmt.Print(str.contents)
	return
}

func nativePrintInteger(f *frame) {
	i := int32(f.variables[0].(stackInt32))
	fmt.Print(i)
	return
}

func (vm *VM) execute(class class, methodName string, previousFrame *frame) {
	method := class.getMethod(methodName)
	var frame frame
	if (method.accessFlags & Native) != 0 {
		frame.variables = make([]stackItem, method.numArgs())
	} else {
		frame.variables = make([]stackItem, method.code.maxLocals)
	}
	for i := method.numArgs() - 1; i >= 0; i-- {
		frame.variables[i] = previousFrame.stack.pop()
	}

	if (method.accessFlags & Native) != 0 {
		native := vm.nativeMethods[methodName]
		if native == nil {
			log.Panicf("Unknown native method %s", methodName)
		}
		native(&frame)
		return
	}

	pc := 0
	for {
		instruction := method.code.code[pc]
		pc++
		switch instruction {
		case 0:
			break
		case 5:
			frame.stack.pushInt32(2)
			break
		case 8:
			frame.stack.pushInt32(5)
			break
		case 16:
			frame.stack.pushInt32(int32(method.code.code[pc]))
			pc++
		case 18:
			strRef := class.constantPoolItems[method.code.code[pc]-1].(stringConstant)
			str := class.constantPoolItems[strRef.utf8Index-1].(utf8String)
			frame.stack.pushString(str)
			pc++
			break
		case 26:
			frame.stack.pushInt32(int32(frame.variables[0].(stackInt32)))
		case 27:
			frame.stack.pushInt32(int32(frame.variables[1].(stackInt32)))
		case 60:
			frame.variables[1] = frame.stack.pop()
		case 96:
			//TODO: make sure we do overflow correctly
			x := frame.stack.popInt32()
			y := frame.stack.popInt32()
			frame.stack.pushInt32(x + y)
		case 172:
			previousFrame.stack.push(frame.stack.pop())
			return
		case 177:
			return
		case 184:
			var i uint16
			i = uint16(method.code.code[pc]) << 8
			pc++
			i |= uint16(method.code.code[pc])
			pc++
			m := class.constantPoolItems[i-1].(methodRef)
			nt := class.constantPoolItems[m.nameAndTypeIndex-1].(nameAndType)
			n := class.constantPoolItems[nt.nameIndex-1].(utf8String).contents
			vm.execute(class, n, &frame)
			break
		default:
			panic(fmt.Sprintf("Unknown instruction: %v", instruction))
		}
	}
}

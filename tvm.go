package main

import (
	"flag"
	"fmt"
	"io/ioutil"
)

func (c *RawClass) getMethod(name string) Method {
	for _, m := range c.methods {
		n := c.constantPoolItems[m.nameIndex-1].(UTF8String).Contents
		if n == name {
			return m
		}
	}
	panic(fmt.Sprintf("Could not find method called %v", name))
}

func (class *RawClass) execute(methodName string, stack []byte) {
	method := class.getMethod(methodName)

	if (method.accessFlags&Native) != 0 && methodName == "print" {
		index := class.constantPoolItems[stack[len(stack)-1]-1].(StringConstant).UTF8Index
		fmt.Print(class.constantPoolItems[index-1].(UTF8String).Contents)
		return
	}

	pc := 0
	for {
		instruction := method.code.code[pc]
		pc += 1
		if instruction == 18 {
			stack = append(stack, method.code.code[pc])
			pc += 1
		} else if instruction == 184 {
			var i uint16
			i = uint16(method.code.code[pc]) << 8
			pc += 1
			i |= uint16(method.code.code[pc])
			pc += 1
			m := class.constantPoolItems[i-1].(MethodRef)
			nt := class.constantPoolItems[m.NameAndTypeIndex-1].(NameAndType)
			n := class.constantPoolItems[nt.NameIndex-1].(UTF8String).Contents
			class.execute(n, stack)
		} else if instruction == 177 {
			break
		} else {
			panic(fmt.Sprintf("Unknow instruction: %v", instruction))
		}
	}
}

func main() {
	flag.Parse()
	filename := flag.Arg(0)
	bytes, _ := ioutil.ReadFile(filename)
	class, _ := parse(bytes)
	var stack []byte
	class.execute("main", stack)
}

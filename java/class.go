package java

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type ConstantPoolItem interface {
	isConstantPoolItem()
}

type AccessFlags uint16

const (
	Public     AccessFlags = 0x0001
	Static                 = 0x0008
	Final                  = 0x0010
	Super                  = 0x0020
	Native                 = 0x0100
	Interface              = 0x0200
	Abstract               = 0x0400
	Synthetic              = 0x1000
	Annotation             = 0x2000
	Enum                   = 0x4000
)

type Code struct {
	maxStack  uint16
	maxLocals uint16
	code      []byte
}

type Method struct {
	accessFlags     AccessFlags
	nameIndex       uint16
	descriptorIndex uint16
	code            Code
}

type RawClass struct {
	magic             uint32
	minorVersion      uint16
	majorVersion      uint16
	constantPoolItems []ConstantPoolItem
	accessFlags       AccessFlags
	thisClass         uint16
	superClass        uint16
	interfaces        []uint16
	fields            []uint16
	methods           []Method
}

type Class struct {
	Name string
}

func fromRawClass(raw RawClass) (class Class) {
	classInfo := raw.constantPoolItems[raw.thisClass-1].(ClassInfo)
	class.Name = raw.constantPoolItems[classInfo.NameIndex-1].(UTF8String).Contents
	return
}

func parseCode(buf *bytes.Reader, length uint32) (c Code) {
	binary.Read(buf, binary.BigEndian, &c.maxStack)
	binary.Read(buf, binary.BigEndian, &c.maxLocals)
	var codeLength uint32
	binary.Read(buf, binary.BigEndian, &codeLength)
	c.code = make([]byte, codeLength)
	for k := 0; k < len(c.code); k++ {
		binary.Read(buf, binary.BigEndian, &c.code[k])
	}
	for k := uint32(8) + codeLength; k < length; k++ {
		var bytes uint8
		if binary.Read(buf, binary.BigEndian, &bytes) != nil {
			panic("Error reading code block")
		}
	}
	return
}

func ParseClass(b []byte) (c RawClass, err error) {
	buf := bytes.NewReader(b)
	err = binary.Read(buf, binary.BigEndian, &c.magic)
	if err != nil {
		return
	}
	err = binary.Read(buf, binary.BigEndian, &c.minorVersion)
	if err != nil {
		return
	}
	err = binary.Read(buf, binary.BigEndian, &c.majorVersion)
	if err != nil {
		return
	}
	var constantPoolCount uint16
	err = binary.Read(buf, binary.BigEndian, &constantPoolCount)
	if err != nil {
		return
	}
	constantPoolCount -= 1
	c.constantPoolItems = make([]ConstantPoolItem, constantPoolCount)
	var i uint16
	for i = 0; i < constantPoolCount; i++ {
		c.constantPoolItems[i] = parseConstantPoolItem(buf)
	}
	err = binary.Read(buf, binary.BigEndian, &c.accessFlags)
	if err != nil {
		return
	}
	err = binary.Read(buf, binary.BigEndian, &c.thisClass)
	if err != nil {
		return
	}
	err = binary.Read(buf, binary.BigEndian, &c.superClass)
	if err != nil {
		return
	}
	var interfacesCount uint16
	err = binary.Read(buf, binary.BigEndian, &interfacesCount)
	if err != nil {
		return
	}
	c.interfaces = make([]uint16, interfacesCount)

	var fieldsCount uint16
	err = binary.Read(buf, binary.BigEndian, &fieldsCount)
	if err != nil {
		return
	}
	c.fields = make([]uint16, fieldsCount)

	var methodsCount uint16
	err = binary.Read(buf, binary.BigEndian, &methodsCount)
	if err != nil {
		return
	}
	c.methods = make([]Method, methodsCount)
	for i = 0; i < methodsCount; i++ {
		err = binary.Read(buf, binary.BigEndian, &c.methods[i].accessFlags)
		err = binary.Read(buf, binary.BigEndian, &c.methods[i].nameIndex)
		err = binary.Read(buf, binary.BigEndian, &c.methods[i].descriptorIndex)
		var attrCount uint16
		err = binary.Read(buf, binary.BigEndian, &attrCount)
		if err != nil {
			return
		}
		for j := uint16(0); j < attrCount; j++ {
			var name uint16
			var length uint32
			err = binary.Read(buf, binary.BigEndian, &name)
			err = binary.Read(buf, binary.BigEndian, &length)
			actualName := (c.constantPoolItems[name-1]).(UTF8String)
			if actualName.Contents == "Code" {
				c.methods[i].code = parseCode(buf, length)
			} else {
				for k := uint32(0); k < length; k++ {
					var bytes uint8
					err = binary.Read(buf, binary.BigEndian, &bytes)
				}
			}
		}
	}
	var attrCount uint16
	err = binary.Read(buf, binary.BigEndian, &attrCount)
	if err != nil {
		return
	}
	for j := uint16(0); j < attrCount; j++ {
		var name uint16
		var length uint32
		err = binary.Read(buf, binary.BigEndian, &name)
		err = binary.Read(buf, binary.BigEndian, &length)
		for k := uint32(0); k < length; k++ {
			var bytes uint8
			err = binary.Read(buf, binary.BigEndian, &bytes)
		}
	}

	return
}

func (c *RawClass) getMethod(name string) Method {
	for _, m := range c.methods {
		n := c.constantPoolItems[m.nameIndex-1].(UTF8String).Contents
		if n == name {
			return m
		}
	}
	panic(fmt.Sprintf("Could not find method called %v", name))
}

func (class *RawClass) Execute(methodName string, stack []byte) {
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
		switch instruction {
		case 18:
			stack = append(stack, method.code.code[pc])
			pc += 1
			break
		case 184:
			var i uint16
			i = uint16(method.code.code[pc]) << 8
			pc += 1
			i |= uint16(method.code.code[pc])
			pc += 1
			m := class.constantPoolItems[i-1].(MethodRef)
			nt := class.constantPoolItems[m.NameAndTypeIndex-1].(NameAndType)
			n := class.constantPoolItems[nt.NameIndex-1].(UTF8String).Contents
			class.Execute(n, stack)
			break
		case 177:
			return
		default:
			panic(fmt.Sprintf("Unknow instruction: %v", instruction))
		}
	}
}

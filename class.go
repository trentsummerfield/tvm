package main

import (
	"bytes"
	"encoding/binary"
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

type NameAndType struct {
	NameIndex       uint16
	DescriptorIndex uint16
}

func (_ NameAndType) isConstantPoolItem() {}

type UTF8String struct {
	Contents string
}

func (_ UTF8String) isConstantPoolItem() {}

type ClassInfo struct {
	NameIndex uint16
}

func (_ ClassInfo) isConstantPoolItem() {}

type MethodRef struct {
	ClassIndex       uint16
	NameAndTypeIndex uint16
}

func (_ MethodRef) isConstantPoolItem() {}

type FieldRef struct {
	ClassIndex       uint16
	NameAndTypeIndex uint16
}

func (_ FieldRef) isConstantPoolItem() {}

type StringConstant struct {
	UTF8Index uint16
}

func (_ StringConstant) isConstantPoolItem() {}

func parseMethodRef(buf *bytes.Reader) (m MethodRef, err error) {
	err = binary.Read(buf, binary.BigEndian, &m)
	return
}

func parseFieldRef(buf *bytes.Reader) (f FieldRef, err error) {
	err = binary.Read(buf, binary.BigEndian, &f)
	return
}

func parseStringConstant(buf *bytes.Reader) (s StringConstant, err error) {
	err = binary.Read(buf, binary.BigEndian, &s)
	return
}

func parseClassInfo(buf *bytes.Reader) (c ClassInfo, err error) {
	err = binary.Read(buf, binary.BigEndian, &c)
	return
}

func parseNameAndType(buf *bytes.Reader) (n NameAndType, err error) {
	err = binary.Read(buf, binary.BigEndian, &n)
	return
}

func parseUTF8String(buf *bytes.Reader) (s UTF8String, err error) {
	var length uint16
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		return
	}
	bytes := make([]byte, length)
	var i uint16
	for i = 0; i < length; i++ {
		bytes[i], err = buf.ReadByte()
		if err != nil {
			return
		}
	}
	s.Contents = string(bytes)
	return
}

type ConstantPoolTag uint8

const (
	CP_UTF8String  ConstantPoolTag = 1
	CP_ClassInfo                   = 7
	CP_String                      = 8
	CP_Field                       = 9
	CP_Method                      = 10
	CP_NameAndType                 = 12
)

func parseConstantPoolItem(buf *bytes.Reader) (ConstantPoolItem, error) {
	var tag ConstantPoolTag
	err := binary.Read(buf, binary.BigEndian, &tag)
	if err != nil {
		panic("Could not read constant pool tag")
	}
	switch tag {
	case CP_UTF8String:
		return parseUTF8String(buf)
	case CP_ClassInfo:
		return parseClassInfo(buf)
	case CP_String:
		return parseStringConstant(buf)
	case CP_Field:
		return parseFieldRef(buf)
	case CP_Method:
		return parseMethodRef(buf)
	case CP_NameAndType:
		return parseNameAndType(buf)
	default:
		panic("Unknown constant pool item")
	}
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

func parse(b []byte) (c RawClass, err error) {
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
		c.constantPoolItems[i], err = parseConstantPoolItem(buf)
		if err != nil {
			return
		}
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

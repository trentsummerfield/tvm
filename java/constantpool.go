package java

import (
	"bytes"
	"encoding/binary"
)

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

func parseMethodRef(buf *bytes.Reader) ConstantPoolItem {
	var m MethodRef
	binary.Read(buf, binary.BigEndian, &m)
	return m
}

func parseFieldRef(buf *bytes.Reader) ConstantPoolItem {
	var f FieldRef
	binary.Read(buf, binary.BigEndian, &f)
	return f
}

func parseStringConstant(buf *bytes.Reader) ConstantPoolItem {
	var s StringConstant
	binary.Read(buf, binary.BigEndian, &s)
	return s
}

func parseClassInfo(buf *bytes.Reader) ConstantPoolItem {
	var c ClassInfo
	binary.Read(buf, binary.BigEndian, &c)
	return c
}

func parseNameAndType(buf *bytes.Reader) ConstantPoolItem {
	var n NameAndType
	binary.Read(buf, binary.BigEndian, &n)
	return n
}

func parseUTF8String(buf *bytes.Reader) ConstantPoolItem {
	var length uint16
	err := binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		panic("Could not parse UTF8 String")
	}
	bytes := make([]byte, length)
	var i uint16
	for i = 0; i < length; i++ {
		bytes[i], err = buf.ReadByte()
		if err != nil {
			panic("Could not parse UTF8 String")
		}
	}
	return UTF8String{string(bytes)}
}

func unknownConstantPoolItem(_ *bytes.Reader) ConstantPoolItem {
	panic("Unknown constant pool item")
}

func parseConstantPoolItem(buf *bytes.Reader) ConstantPoolItem {
	parsers := []func(*bytes.Reader) ConstantPoolItem{
		unknownConstantPoolItem,
		parseUTF8String,
		unknownConstantPoolItem,
		unknownConstantPoolItem,
		unknownConstantPoolItem,
		unknownConstantPoolItem,
		unknownConstantPoolItem,
		parseClassInfo,
		parseStringConstant,
		parseFieldRef,
		parseMethodRef,
		unknownConstantPoolItem,
		parseNameAndType,
	}
	var tag uint8
	err := binary.Read(buf, binary.BigEndian, &tag)
	if err != nil {
		panic("Could not read constant pool tag")
	}
	return parsers[tag](buf)
}

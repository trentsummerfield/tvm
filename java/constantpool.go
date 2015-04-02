package java

import (
	"bytes"
	"encoding/binary"
)

type nameAndType struct {
	nameIndex       uint16
	descriptorIndex uint16
}

func (_ nameAndType) isConstantPoolItem() {}

func parseNameAndType(buf *bytes.Reader) constantPoolItem {
	var nameIndex, descriptorIndex uint16
	binary.Read(buf, binary.BigEndian, &nameIndex)
	binary.Read(buf, binary.BigEndian, &descriptorIndex)
	return nameAndType{nameIndex, descriptorIndex}
}

type utf8String struct {
	contents string
}

func (_ utf8String) isConstantPoolItem() {}

func parseUTF8String(buf *bytes.Reader) constantPoolItem {
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
	return utf8String{string(bytes)}
}

type classInfo struct {
	nameIndex uint16
}

func (_ classInfo) isConstantPoolItem() {}

func parseClassInfo(buf *bytes.Reader) constantPoolItem {
	var nameIndex uint16
	binary.Read(buf, binary.BigEndian, &nameIndex)
	return classInfo{nameIndex}
}

type methodRef struct {
	classIndex       uint16
	nameAndTypeIndex uint16
}

func (_ methodRef) isConstantPoolItem() {}

func parseMethodRef(buf *bytes.Reader) constantPoolItem {
	var classIndex, nameAndTypeIndex uint16
	binary.Read(buf, binary.BigEndian, &classIndex)
	binary.Read(buf, binary.BigEndian, &nameAndTypeIndex)
	return methodRef{classIndex, nameAndTypeIndex}
}

type fieldRef struct {
	classIndex       uint16
	nameAndTypeIndex uint16
}

func (_ fieldRef) isConstantPoolItem() {}

func parseFieldRef(buf *bytes.Reader) constantPoolItem {
	var classIndex, nameAndTypeIndex uint16
	binary.Read(buf, binary.BigEndian, &classIndex)
	binary.Read(buf, binary.BigEndian, &nameAndTypeIndex)
	return fieldRef{classIndex, nameAndTypeIndex}
}

type stringConstant struct {
	utf8Index uint16
}

func (_ stringConstant) isConstantPoolItem() {}

func parseStringConstant(buf *bytes.Reader) constantPoolItem {
	var utf8Index uint16
	binary.Read(buf, binary.BigEndian, &utf8Index)
	return stringConstant{utf8Index}
}

func unknownConstantPoolItem(_ *bytes.Reader) constantPoolItem {
	panic("Unknown constant pool item")
}

func parseConstantPoolItem(buf *bytes.Reader) constantPoolItem {
	parsers := []func(*bytes.Reader) constantPoolItem{
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

package main

import (
	"bytes"
	"encoding/binary"
)

type ConstantPoolItem interface {
	isConstantPoolItem()
}

type RawClass struct {
	magic             uint32
	minorVersion      uint16
	majorVersion      uint16
	constantPoolItems []ConstantPoolItem
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

type ParseError struct {
	reason string
}

func (e ParseError) Error() string {
	return e.reason
}

func parseMethodRef(buf *bytes.Reader) (m MethodRef, err error) {
	err = binary.Read(buf, binary.BigEndian, &m)
	return
}

func parseFieldRef(buf *bytes.Reader) (m FieldRef, err error) {
	err = binary.Read(buf, binary.BigEndian, &m)
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

func parseNameAndType(buf *bytes.Reader) (c NameAndType, err error) {
	err = binary.Read(buf, binary.BigEndian, &c)
	return
}

func parseUTF8String(buf *bytes.Reader) (c UTF8String, err error) {
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
	c.Contents = string(bytes)
	return
}

func parseConstantPoolItem(buf *bytes.Reader) (ConstantPoolItem, error) {
	tag, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	if tag == 10 {
		return parseMethodRef(buf)
	} else if tag == 9 {
		return parseFieldRef(buf)
	} else if tag == 8 {
		return parseStringConstant(buf)
	} else if tag == 7 {
		return parseClassInfo(buf)
	} else if tag == 1 {
		return parseUTF8String(buf)
	} else if tag == 12 {
		return parseNameAndType(buf)
	}
	panic("Unknown constant pool item")
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
	return
}

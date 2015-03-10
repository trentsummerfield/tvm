package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Class struct {
	magic             uint32
	minorVersion      uint16
	majorVersion      uint16
	constantPoolCount uint16
	constantPoolItems []interface{}
}

type NameAndType struct {
	NameIndex       uint16
	DescriptorIndex uint16
}

func (m NameAndType) String() string {
	return fmt.Sprintf("NameAndType: %v, %v", m.NameIndex, m.DescriptorIndex)
}

type UTF8String struct {
	Length uint16
	Bytes  []uint8
}

func (m UTF8String) String() string {
	return fmt.Sprintf("UTF8: %v", string(m.Bytes))
}

type ClassInfo struct {
	NameIndex uint16
}

func (m ClassInfo) String() string {
	return fmt.Sprintf("Class: %v", m.NameIndex)
}

type MethodRef struct {
	ClassIndex       uint16
	NameAndTypeIndex uint16
}

func (m MethodRef) String() string {
	return fmt.Sprintf("Method: %v, %v", m.ClassIndex, m.NameAndTypeIndex)
}

type FieldRef struct {
	ClassIndex       uint16
	NameAndTypeIndex uint16
}

func (m FieldRef) String() string {
	return fmt.Sprintf("Field: %v, %v", m.ClassIndex, m.NameAndTypeIndex)
}

type StringConstant struct {
	UTF8Index uint16
}

func (m StringConstant) String() string {
	return fmt.Sprintf("String: %v", m.UTF8Index)
}

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
	err = binary.Read(buf, binary.BigEndian, &c.Length)
	if err != nil {
		return
	}
	c.Bytes = make([]byte, c.Length)
	var i uint16
	for i = 0; i < c.Length; i++ {
		c.Bytes[i], err = buf.ReadByte()
		if err != nil {
			return
		}
	}
	return
}

func parseConstantPoolItem(buf *bytes.Reader) (interface{}, error) {
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
	return tag, nil
}

func parse(b []byte) (c Class, err error) {
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
	err = binary.Read(buf, binary.BigEndian, &c.constantPoolCount)
	if err != nil {
		return
	}
	c.constantPoolItems = make([]interface{}, c.constantPoolCount)
	var i uint16
	for i = 0; i < c.constantPoolCount; i++ {
		c.constantPoolItems[i], err = parseConstantPoolItem(buf)
		if err != nil {
			return
		}
	}
	return
}

package java

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
)

type constantPoolItem interface {
	isConstantPoolItem()
}

type accessFlags uint16

const (
	Public     accessFlags = 0x0001
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

type code struct {
	maxStack  uint16
	maxLocals uint16
	code      []byte
}

type class struct {
	magic             uint32
	minorVersion      uint16
	majorVersion      uint16
	constantPoolItems []constantPoolItem
	accessFlags       accessFlags
	thisClass         uint16
	superClass        uint16
	interfaces        []uint16
	fields            []field
	methods           []method
	initialised       bool
}

func parseCode(cr classDecoder, length uint32) (c code) {
	c.maxStack = cr.u2()
	c.maxLocals = cr.u2()
	codeLength := cr.u4()
	c.code = make([]byte, codeLength)
	for k := 0; k < len(c.code); k++ {
		c.code[k] = cr.u1()
	}
	for k := uint32(8) + codeLength; k < length; k++ {
		_ = cr.u1()
	}
	return
}

type classDecoder struct {
	reader io.Reader
	err    error
}

func (r classDecoder) u4() uint32 {
	if r.err != nil {
		return 0
	}
	var x uint32
	r.err = binary.Read(r.reader, binary.BigEndian, &x)
	return x
}

func (r classDecoder) u2() uint16 {
	if r.err != nil {
		return 0
	}
	var x uint16
	r.err = binary.Read(r.reader, binary.BigEndian, &x)
	return x
}

func (r classDecoder) u1() uint8 {
	if r.err != nil {
		return 0
	}
	var x uint8
	r.err = binary.Read(r.reader, binary.BigEndian, &x)
	return x
}

func newClassDecoder(r io.Reader) classDecoder {
	cr := classDecoder{r, nil}
	magic := cr.u4()
	if magic != 0xCAFEBABE {
		cr.err = errors.New("Bad magic number")
	}
	return cr
}

func parseClass(r io.Reader) (c class, err error) {
	cr := newClassDecoder(r)
	_ = cr.u2() // minor version
	_ = cr.u2() // major version
	cpc := cr.u2()
	constantPoolCount := cpc - 1
	if cpc != 0 {
		c.constantPoolItems = make([]constantPoolItem, constantPoolCount)
		for i := uint16(0); i < constantPoolCount; i++ {
			c.constantPoolItems[i] = parseConstantPoolItem(&c, cr)
		}
	}

	c.accessFlags = accessFlags(cr.u2())
	c.thisClass = cr.u2()
	c.superClass = cr.u2()

	interfacesCount := cr.u2()
	c.interfaces = make([]uint16, interfacesCount)

	fieldsCount := cr.u2()
	c.fields = make([]field, fieldsCount)
	for i := uint16(0); i < fieldsCount; i++ {
		c.fields[i].accessFlags = accessFlags(cr.u2())
		c.fields[i].nameIndex = cr.u2()
		c.fields[i].descriptorIndex = cr.u2()

		attrCount := cr.u2()
		for j := uint16(0); j < attrCount; j++ {
			_ = cr.u2()
			length := cr.u4()
			for k := uint32(0); k < length; k++ {
				_ = cr.u1() // throw away bytes
			}
		}
	}

	methodsCount := cr.u2()
	c.methods = make([]method, methodsCount)
	for i := uint16(0); i < methodsCount; i++ {
		c.methods[i].class = c
		c.methods[i].accessFlags = accessFlags(cr.u2())
		c.methods[i].nameIndex = cr.u2()
		c.methods[i].descriptorIndex = cr.u2()

		var sig string
		sig = c.constantPoolItems[c.methods[i].descriptorIndex-1].(utf8String).contents
		c.methods[i].signiture = parseSigniture(sig)

		attrCount := cr.u2()
		for j := uint16(0); j < attrCount; j++ {
			name := cr.u2()
			length := cr.u4()
			actualName := (c.constantPoolItems[name-1]).(utf8String)
			if actualName.contents == "Code" {
				c.methods[i].code = parseCode(cr, length)
			} else {
				for k := uint32(0); k < length; k++ {
					_ = cr.u1() // throw away bytes
				}
			}
		}
	}
	attrCount := cr.u2()
	for j := uint16(0); j < attrCount; j++ {
		_ = cr.u2()
		length := cr.u4()
		for k := uint32(0); k < length; k++ {
			_ = cr.u1() // throw away bytes
		}
	}

	return c, cr.err
}

func (c *class) hasMethodCalled(name string) bool {
	for _, m := range c.methods {
		n := c.constantPoolItems[m.nameIndex-1].(utf8String).contents
		if n == name {
			return true
		}
	}
	return false
}

func (c *class) getMethod(name string) method {
	for _, m := range c.methods {
		n := c.constantPoolItems[m.nameIndex-1].(utf8String).contents
		if n == name {
			return m
		}
	}
	panic(fmt.Sprintf("Could not find method called %v", name))
}

func (c *class) getField(name string) *field {
	for i, f := range c.fields {
		n := c.constantPoolItems[f.nameIndex-1].(utf8String).contents
		if n == name {
			return &(c.fields[i])
		}
	}
	panic(fmt.Sprintf("Could not find field called %v", name))
}

func (c *class) getName() string {
	info := c.constantPoolItems[c.thisClass-1].(classInfo)
	name := c.constantPoolItems[info.nameIndex-1].(utf8String)
	return name.contents
}

func (c *class) getSuperName() string {
	info := c.constantPoolItems[c.superClass-1].(classInfo)
	name := c.constantPoolItems[info.nameIndex-1].(utf8String)
	return name.contents
}

func (c *class) getMethodRefAt(index uint16) methodRef {
	return c.constantPoolItems[index-1].(methodRef)
}

func (c *class) getFieldRefAt(index uint16) fieldRef {
	return c.constantPoolItems[index-1].(fieldRef)
}

func (c *class) getClassInfoAt(index uint16) classInfo {
	return c.constantPoolItems[index-1].(classInfo)
}

func (c *class) getStringAt(index int) utf8String {
	strRef := c.constantPoolItems[index].(stringConstant)
	return c.constantPoolItems[strRef.utf8Index-1].(utf8String)
}

func (m methodRef) methodName() string {
	nt := m.containingClass.constantPoolItems[m.nameAndTypeIndex-1].(nameAndType)
	n := m.containingClass.constantPoolItems[nt.nameIndex-1].(utf8String).contents
	return n
}

func (m methodRef) className() string {
	ct := m.containingClass.constantPoolItems[m.classIndex-1].(classInfo)
	c := m.containingClass.constantPoolItems[ct.nameIndex-1].(utf8String).contents
	return c
}

func (ct classInfo) className() string {
	c := ct.containingClass.constantPoolItems[ct.nameIndex-1].(utf8String).contents
	return c
}

func (m fieldRef) fieldName() string {
	nt := m.containingClass.constantPoolItems[m.nameAndTypeIndex-1].(nameAndType)
	n := m.containingClass.constantPoolItems[nt.nameIndex-1].(utf8String).contents
	return n
}

func (m fieldRef) className() string {
	ct := m.containingClass.constantPoolItems[m.classIndex-1].(classInfo)
	c := m.containingClass.constantPoolItems[ct.nameIndex-1].(utf8String).contents
	return c
}

func parseSigniture(sig string) []string {
	s := make([]string, 0)
	className := false
	for _, c := range sig {
		//TODO: save the name of the class. Maybe
		if className {
			if c == ';' {
				className = false
			}
			continue
		}
		switch c {
		case '(', ')', '[':
			break
		case 'B':
			s = append(s, "byte")
			break
		case 'C':
			s = append(s, "char")
			break
		case 'D':
			s = append(s, "double")
			break
		case 'F':
			s = append(s, "float")
			break
		case 'I':
			s = append(s, "int")
			break
		case 'J':
			s = append(s, "long")
			break
		case 'S':
			s = append(s, "short")
			break
		case 'Z':
			s = append(s, "boolean")
			break
		case 'V':
			s = append(s, "void")
			break
		case 'L':
			s = append(s, "reference")
			className = true
			break
		default:
			log.Panicf("Can't parse signiture: %s", sig)
		}
	}
	return s
}

type nameAndType struct {
	nameIndex       uint16
	descriptorIndex uint16
}

func (_ nameAndType) isConstantPoolItem() {}

func parseNameAndType(c *class, cr classDecoder) constantPoolItem {
	nameIndex := cr.u2()
	descriptorIndex := cr.u2()
	return nameAndType{nameIndex, descriptorIndex}
}

type utf8String struct {
	contents string
}

func (_ utf8String) isConstantPoolItem() {}

func parseUTF8String(c *class, cr classDecoder) constantPoolItem {
	length := cr.u2()
	bytes := make([]byte, length)
	for i := uint16(0); i < length; i++ {
		bytes[i] = cr.u1()
	}
	return utf8String{string(bytes)}
}

type classInfo struct {
	containingClass *class
	nameIndex       uint16
}

func (_ classInfo) isConstantPoolItem() {}

func parseClassInfo(c *class, cr classDecoder) constantPoolItem {
	nameIndex := cr.u2()
	return classInfo{c, nameIndex}
}

type methodRef struct {
	containingClass  *class
	classIndex       uint16
	nameAndTypeIndex uint16
}

func (_ methodRef) isConstantPoolItem() {}

func parseMethodRef(c *class, cr classDecoder) constantPoolItem {
	classIndex := cr.u2()
	nameAndTypeIndex := cr.u2()
	return methodRef{c, classIndex, nameAndTypeIndex}
}

type fieldRef struct {
	containingClass  *class
	classIndex       uint16
	nameAndTypeIndex uint16
}

func (_ fieldRef) isConstantPoolItem() {}

func parseFieldRef(c *class, cr classDecoder) constantPoolItem {
	classIndex := cr.u2()
	nameAndTypeIndex := cr.u2()
	return fieldRef{c, classIndex, nameAndTypeIndex}
}

type stringConstant struct {
	utf8Index uint16
}

func (_ stringConstant) isConstantPoolItem() {}

func parseStringConstant(c *class, cr classDecoder) constantPoolItem {
	utf8Index := cr.u2()
	return stringConstant{utf8Index}
}

func unknownConstantPoolItem(_ *class, _ classDecoder) constantPoolItem {
	panic("Unknown constant pool item")
}

func parseConstantPoolItem(c *class, cr classDecoder) constantPoolItem {
	parsers := []func(*class, classDecoder) constantPoolItem{
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
	tag := cr.u1()
	return parsers[tag](c, cr)
}

type field struct {
	accessFlags     accessFlags
	nameIndex       uint16
	descriptorIndex uint16
	value           javaValue
}

type method struct {
	class           class
	signiture       []string
	accessFlags     accessFlags
	nameIndex       uint16
	descriptorIndex uint16
	code            code
}

func (m method) name() string {
	return m.class.constantPoolItems[m.nameIndex-1].(utf8String).contents
}

func (m method) numArgs() int {
	return len(m.signiture) - 1
}

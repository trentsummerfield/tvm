package java

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
)

type ConstantPoolItem interface {
	isConstantPoolItem()
	String() string
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

type Code struct {
	maxStack     uint16
	maxLocals    uint16
	Instructions []byte
}

type Class struct {
	magic             uint32
	minorVersion      uint16
	majorVersion      uint16
	ConstantPoolItems []ConstantPoolItem
	accessFlags       accessFlags
	thisClass         uint16
	superClass        uint16
	interfaces        []uint16
	fields            []field
	methods           []Method
	initialised       bool
}

func parseCode(cr classDecoder, length uint32) (c Code) {
	c.maxStack = cr.u2()
	c.maxLocals = cr.u2()
	codeLength := cr.u4()
	c.Instructions = make([]byte, codeLength)
	for k := 0; k < len(c.Instructions); k++ {
		c.Instructions[k] = cr.u1()
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

func parseConstantPool(c *Class, cr classDecoder, constantPoolCount uint16) []ConstantPoolItem {
	items := make([]ConstantPoolItem, constantPoolCount)
	for i := uint16(0); i < constantPoolCount; i++ {
		tag := cr.u1()
		switch tag {
		case 1:
			items[i] = parseUTF8String(c, cr)
		case 5:
			items[i] = parseLongConstant(c, cr)
			i++
		case 7:
			items[i] = parseClassInfo(c, cr)
		case 8:
			items[i] = parseStringConstant(c, cr)
		case 9:
			items[i] = parseFieldRef(c, cr)
		case 10:
			items[i] = parseMethodRef(c, cr)
		case 12:
			items[i] = parseNameAndType(c, cr)
		default:
			log.Fatalf("Unknown tag %d\n", tag)
		}
	}
	return items
}

func parseClass(r io.Reader) (c Class, err error) {
	cr := newClassDecoder(r)
	_ = cr.u2() // minor version
	_ = cr.u2() // major version
	cpc := cr.u2()
	constantPoolCount := cpc - 1
	if cpc != 0 {
		c.ConstantPoolItems = parseConstantPool(&c, cr, constantPoolCount)
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
	c.methods = make([]Method, methodsCount)
	for i := uint16(0); i < methodsCount; i++ {
		c.methods[i].class = &c
		c.methods[i].accessFlags = accessFlags(cr.u2())
		c.methods[i].nameIndex = cr.u2()
		c.methods[i].descriptorIndex = cr.u2()

		var sig string
		sig = c.ConstantPoolItems[c.methods[i].descriptorIndex-1].(utf8String).contents
		c.methods[i].signiture = parseSigniture(sig)

		attrCount := cr.u2()
		for j := uint16(0); j < attrCount; j++ {
			name := cr.u2()
			length := cr.u4()
			actualName := (c.ConstantPoolItems[name-1]).(utf8String)
			if actualName.contents == "Code" {
				c.methods[i].Code = parseCode(cr, length)
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

func (c *Class) hasMethodCalled(name string) bool {
	for _, m := range c.methods {
		n := c.ConstantPoolItems[m.nameIndex-1].(utf8String).contents
		if n == name {
			return true
		}
	}
	return false
}

func (c *Class) resolveMethod(name string) Method {
	for _, m := range c.methods {
		n := c.ConstantPoolItems[m.nameIndex-1].(utf8String).contents
		if n == name {
			return m
		}
	}
	panic(fmt.Sprintf("Could not find method called %v", name))
}

func (c *Class) getField(name string) *field {
	for i, f := range c.fields {
		n := c.ConstantPoolItems[f.nameIndex-1].(utf8String).contents
		if n == name {
			return &(c.fields[i])
		}
	}
	panic(fmt.Sprintf("Could not find field called %v", name))
}

func (c *Class) Name() string {
	info := c.ConstantPoolItems[c.thisClass-1].(classInfo)
	name := c.ConstantPoolItems[info.nameIndex-1].(utf8String)
	return name.contents
}

func (c *Class) getSuperName() string {
	info := c.ConstantPoolItems[c.superClass-1].(classInfo)
	name := c.ConstantPoolItems[info.nameIndex-1].(utf8String)
	return name.contents
}

func (c *Class) getMethodRefAt(index uint16) methodRef {
	return c.ConstantPoolItems[index-1].(methodRef)
}

func (c *Class) getFieldRefAt(index uint16) fieldRef {
	return c.ConstantPoolItems[index-1].(fieldRef)
}

func (c *Class) getClassInfoAt(index uint16) classInfo {
	return c.ConstantPoolItems[index-1].(classInfo)
}

func (c *Class) getStringAt(index int) utf8String {
	strRef := c.ConstantPoolItems[index].(stringConstant)
	return c.ConstantPoolItems[strRef.utf8Index-1].(utf8String)
}

func (c *Class) getLongAt(index int) longConstant {
	return c.ConstantPoolItems[index].(longConstant)
}

func (m methodRef) methodName() string {
	nt := m.containingClass.ConstantPoolItems[m.nameAndTypeIndex-1].(nameAndType)
	n := m.containingClass.ConstantPoolItems[nt.nameIndex-1].(utf8String).contents
	return n
}

func (m methodRef) className() string {
	ct := m.containingClass.ConstantPoolItems[m.classIndex-1].(classInfo)
	c := m.containingClass.ConstantPoolItems[ct.nameIndex-1].(utf8String).contents
	return c
}

func (ct classInfo) className() string {
	c := ct.containingClass.ConstantPoolItems[ct.nameIndex-1].(utf8String).contents
	return c
}

func (m fieldRef) fieldName() string {
	nt := m.containingClass.ConstantPoolItems[m.nameAndTypeIndex-1].(nameAndType)
	n := m.containingClass.ConstantPoolItems[nt.nameIndex-1].(utf8String).contents
	return n
}

func (m fieldRef) className() string {
	ct := m.containingClass.ConstantPoolItems[m.classIndex-1].(classInfo)
	c := m.containingClass.ConstantPoolItems[ct.nameIndex-1].(utf8String).contents
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

func (n nameAndType) String() string {
	return fmt.Sprintf("(NameAndType) name: %d, type: %d", n.nameIndex, n.descriptorIndex)
}

func parseNameAndType(c *Class, cr classDecoder) ConstantPoolItem {
	nameIndex := cr.u2()
	descriptorIndex := cr.u2()
	return nameAndType{nameIndex, descriptorIndex}
}

type utf8String struct {
	contents string
}

func (_ utf8String) isConstantPoolItem() {}

func (u utf8String) String() string {
	return "(String) \"" + u.contents + "\""
}

func parseUTF8String(c *Class, cr classDecoder) ConstantPoolItem {
	length := cr.u2()
	bytes := make([]byte, length)
	for i := uint16(0); i < length; i++ {
		bytes[i] = cr.u1()
	}
	return utf8String{string(bytes)}
}

type classInfo struct {
	containingClass *Class
	nameIndex       uint16
}

func (_ classInfo) isConstantPoolItem() {}

func (c classInfo) String() string {
	return fmt.Sprintf("(ClassInfo) %d", c.nameIndex)
}

func parseClassInfo(c *Class, cr classDecoder) ConstantPoolItem {
	nameIndex := cr.u2()
	return classInfo{c, nameIndex}
}

type methodRef struct {
	containingClass  *Class
	classIndex       uint16
	nameAndTypeIndex uint16
}

func (_ methodRef) isConstantPoolItem() {}

func (m methodRef) String() string {
	return fmt.Sprintf("(MethodRef) class: %d, name: %d", m.classIndex, m.nameAndTypeIndex)
}

func parseMethodRef(c *Class, cr classDecoder) ConstantPoolItem {
	classIndex := cr.u2()
	nameAndTypeIndex := cr.u2()
	return methodRef{c, classIndex, nameAndTypeIndex}
}

type fieldRef struct {
	containingClass  *Class
	classIndex       uint16
	nameAndTypeIndex uint16
}

func (_ fieldRef) isConstantPoolItem() {}

func (f fieldRef) String() string {
	return fmt.Sprintf("(FieldRef) class: %d, name %d", f.classIndex, f.nameAndTypeIndex)
}

func parseFieldRef(c *Class, cr classDecoder) ConstantPoolItem {
	classIndex := cr.u2()
	nameAndTypeIndex := cr.u2()
	return fieldRef{c, classIndex, nameAndTypeIndex}
}

type stringConstant struct {
	utf8Index uint16
}

func (_ stringConstant) isConstantPoolItem() {}

func (s stringConstant) String() string {
	return fmt.Sprintf("(StringConst) index: %d", s.utf8Index)
}

func parseStringConstant(c *Class, cr classDecoder) ConstantPoolItem {
	utf8Index := cr.u2()
	return stringConstant{utf8Index}
}

type longConstant struct {
	value int64
}

func (_ longConstant) isConstantPoolItem() {}

func (l longConstant) String() string {
	return fmt.Sprintf("(Long) %d", l.value)
}

func parseLongConstant(c *Class, cr classDecoder) ConstantPoolItem {
	long := int64(cr.u4()) << 32
	long += int64(cr.u4())
	log.Printf("Parsed %v\n", long)
	return longConstant{long}
}

type field struct {
	accessFlags     accessFlags
	nameIndex       uint16
	descriptorIndex uint16
	value           javaValue
}

type Method struct {
	class           *Class
	signiture       []string
	accessFlags     accessFlags
	nameIndex       uint16
	descriptorIndex uint16
	Code            Code
}

func (m *Method) Name() string {
	return m.class.ConstantPoolItems[m.nameIndex-1].(utf8String).contents
}

func (m *Method) Class() *Class {
	return m.class
}

func (m *Method) numArgs() int {
	return len(m.signiture) - 1
}

func (m *Method) Sig() []string {
	return m.signiture
}

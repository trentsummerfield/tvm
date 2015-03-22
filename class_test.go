package main

import (
	"io/ioutil"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type ClassSuite struct {
	class RawClass
}

var _ = Suite(&ClassSuite{})

func (s *ClassSuite) SetUpSuite(c *C) {
	bytes, err := ioutil.ReadFile("Hello.class")
	c.Assert(err, IsNil)
	s.class, err = parse(bytes)
	c.Assert(err, IsNil)
}

func (s *ClassSuite) TestClassMagicAndVersions(c *C) {
	c.Check(s.class.magic, Equals, uint32(0xCAFEBABE))
	c.Check(s.class.majorVersion, Equals, uint16(52))
	c.Check(s.class.minorVersion, Equals, uint16(0))
}

func (s *ClassSuite) TestClassConstantPool(c *C) {
	pool := s.class.constantPoolItems
	// The constant pool count is 1 more than the actual number of elements in the array.
	// Elements are index from 1 to constantPoolCount-1
	// This makes no sense.
	c.Check(len(pool), Equals, 28)
	c.Check(pool[0], Equals, MethodRef{6, 15})
	c.Check(pool[1], Equals, FieldRef{16, 17})
	c.Check(pool[2], Equals, StringConstant{18})
	c.Check(pool[3], Equals, MethodRef{19, 20})
	c.Check(pool[4], Equals, ClassInfo{21})
	c.Check(pool[5], Equals, ClassInfo{22})
	c.Check(pool[6], Equals, UTF8String{"<init>"})
	c.Check(pool[7], Equals, UTF8String{"()V"})
	c.Check(pool[8], Equals, UTF8String{"Code"})
	c.Check(pool[9], Equals, UTF8String{"LineNumberTable"})
	c.Check(pool[10], Equals, UTF8String{"main"})
	c.Check(pool[11], Equals, UTF8String{"([Ljava/lang/String;)V"})
	c.Check(pool[12], Equals, UTF8String{"SourceFile"})
	c.Check(pool[13], Equals, UTF8String{"Hello.java"})
	c.Check(pool[14], Equals, NameAndType{7, 8})
	c.Check(pool[15], Equals, ClassInfo{23})
	c.Check(pool[16], Equals, NameAndType{24, 25})
	c.Check(pool[17], Equals, UTF8String{"Hello, World"})
	c.Check(pool[18], Equals, ClassInfo{26})
	c.Check(pool[19], Equals, NameAndType{27, 28})
	c.Check(pool[20], Equals, UTF8String{"Hello"})
	c.Check(pool[21], Equals, UTF8String{"java/lang/Object"})
	c.Check(pool[22], Equals, UTF8String{"java/lang/System"})
	c.Check(pool[23], Equals, UTF8String{"out"})
	c.Check(pool[24], Equals, UTF8String{"Ljava/io/PrintStream;"})
	c.Check(pool[25], Equals, UTF8String{"java/io/PrintStream"})
	c.Check(pool[26], Equals, UTF8String{"println"})
	c.Check(pool[27], Equals, UTF8String{"(Ljava/lang/String;)V"})
}

func (s *ClassSuite) TestAccessFlags(c *C) {
	c.Check(s.class.accessFlags, Equals, Public|Super)
}

func (s *ClassSuite) TestThisClassAndSuperClass(c *C) {
	c.Check(s.class.thisClass, Equals, uint16(5))
	c.Check(s.class.superClass, Equals, uint16(6))
}

func (s *ClassSuite) TestInterfacesAndFields(c *C) {
	c.Check(len(s.class.interfaces), Equals, 0)
	c.Check(len(s.class.fields), Equals, 0)
}

func (s *ClassSuite) TestMethods(c *C) {
	c.Check(len(s.class.methods), Equals, 2)
	c.Check(s.class.methods[0].accessFlags, Equals, Public)
	c.Check(s.class.methods[0].nameIndex, Equals, uint16(7))
	c.Check(s.class.methods[0].descriptorIndex, Equals, uint16(8))

	c.Check(s.class.methods[1].accessFlags, Equals, Public|Static)
	c.Check(s.class.methods[1].nameIndex, Equals, uint16(11))
	c.Check(s.class.methods[1].descriptorIndex, Equals, uint16(12))
}

func (s *ClassSuite) TestCode(c *C) {
	c.Check(s.class.methods[0].code.maxStack, Equals, uint16(1))
	c.Check(s.class.methods[0].code.maxLocals, Equals, uint16(1))
	c.Check(len(s.class.methods[0].code.code), Equals, 5)

	c.Check(s.class.methods[1].code.maxStack, Equals, uint16(2))
	c.Check(s.class.methods[1].code.maxLocals, Equals, uint16(1))
	c.Check(len(s.class.methods[1].code.code), Equals, 9)
}

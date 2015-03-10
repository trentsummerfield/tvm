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
	class Class
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
	c.Check(s.class.constantPoolCount, Equals, uint16(29))
	c.Check(s.class.constantPoolItems[0], FitsTypeOf, MethodRef{})
	c.Check(s.class.constantPoolItems[1], FitsTypeOf, FieldRef{})
	c.Check(s.class.constantPoolItems[2], FitsTypeOf, StringConstant{})
	c.Check(s.class.constantPoolItems[3], FitsTypeOf, MethodRef{})
	c.Check(s.class.constantPoolItems[4], FitsTypeOf, ClassInfo{})
	c.Check(s.class.constantPoolItems[5], FitsTypeOf, ClassInfo{})
	c.Check(s.class.constantPoolItems[6], FitsTypeOf, UTF8String{})
	c.Check(s.class.constantPoolItems[7], FitsTypeOf, UTF8String{})
	c.Check(s.class.constantPoolItems[8], FitsTypeOf, UTF8String{})
	c.Check(s.class.constantPoolItems[9], FitsTypeOf, UTF8String{})
	c.Check(s.class.constantPoolItems[10], FitsTypeOf, UTF8String{})
	c.Check(s.class.constantPoolItems[11], FitsTypeOf, UTF8String{})
	c.Check(s.class.constantPoolItems[12], FitsTypeOf, UTF8String{})
	c.Check(s.class.constantPoolItems[13], FitsTypeOf, UTF8String{})
	c.Check(s.class.constantPoolItems[14], FitsTypeOf, NameAndType{})
	c.Check(s.class.constantPoolItems[15], FitsTypeOf, ClassInfo{})
	c.Check(s.class.constantPoolItems[16], FitsTypeOf, NameAndType{})
	c.Check(s.class.constantPoolItems[17], FitsTypeOf, UTF8String{})
	c.Check(s.class.constantPoolItems[18], FitsTypeOf, ClassInfo{})
	c.Check(s.class.constantPoolItems[19], FitsTypeOf, NameAndType{})
	c.Check(s.class.constantPoolItems[20], FitsTypeOf, UTF8String{})
	c.Check(s.class.constantPoolItems[21], FitsTypeOf, UTF8String{})
	c.Check(s.class.constantPoolItems[22], FitsTypeOf, UTF8String{})
	c.Check(s.class.constantPoolItems[23], FitsTypeOf, UTF8String{})
	c.Check(s.class.constantPoolItems[24], FitsTypeOf, UTF8String{})
	c.Check(s.class.constantPoolItems[25], FitsTypeOf, UTF8String{})
	c.Check(s.class.constantPoolItems[26], FitsTypeOf, UTF8String{})
	c.Check(s.class.constantPoolItems[27], FitsTypeOf, UTF8String{})
}

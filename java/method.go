package java

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

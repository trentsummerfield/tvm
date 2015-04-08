package java

type stack struct {
	items []javaValue
	size  uint
}

type javaValue interface {
	isJavaValue()
}

func (s *stack) push(e javaValue) {
	s.items = append(s.items, e)
	s.size++
}

func (s *stack) pop() javaValue {
	if s.size == 0 {
		panic("Cannot pop from an empty stack")
	}
	e := s.items[s.size-1]
	s.size--
	s.items = s.items[:s.size]
	return e
}

func (_ utf8String) isJavaValue() {}

func (s *stack) pushString(str utf8String) {
	s.push(str)
}

func (s *stack) popString() utf8String {
	return s.pop().(utf8String)
}

type javaInt int32

func (_ javaInt) isJavaValue() {}

func (s *stack) pushInt32(i int32) {
	s.push(javaInt(i))
}

func (s *stack) popInt32() int32 {
	return int32(s.pop().(javaInt))
}

func (i javaInt) unbox() int32 {
	return int32(i)
}

type javaByte byte

func (_ javaByte) isJavaValue() {}

func (s *stack) pushByte(i byte) {
	s.push(javaByte(i))
}

func (s *stack) popByte() byte {
	return byte(s.pop().(javaByte))
}

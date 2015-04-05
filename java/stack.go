package java

type stack struct {
	items []stackItem
	size  uint
}

type stackItem interface {
	isStackItem()
}

func (s *stack) push(e stackItem) {
	s.items = append(s.items, e)
	s.size++
}

func (s *stack) pop() stackItem {
	if s.size == 0 {
		panic("Cannot pop from an empty stack")
	}
	e := s.items[s.size-1]
	s.size--
	s.items = s.items[:s.size]
	return e
}

func (_ utf8String) isStackItem() {}

func (s *stack) pushString(str utf8String) {
	s.push(str)
}

func (s *stack) popString() utf8String {
	return s.pop().(utf8String)
}

type stackInt32 int32

func (_ stackInt32) isStackItem() {}

func (s *stack) pushInt32(i int32) {
	s.push(stackInt32(i))
}

func (s *stack) popInt32() int32 {
	return int32(s.pop().(stackInt32))
}

type stackByte byte

func (_ stackByte) isStackItem() {}

func (s *stack) pushByte(i byte) {
	s.push(stackByte(i))
}

func (s *stack) popByte() byte {
	return byte(s.pop().(stackByte))
}

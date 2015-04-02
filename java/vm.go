package java

import (
	"io/ioutil"
)

type VM struct {
	classes []RawClass
}

func NewVM() (vm VM) {
	vm.classes = make([]RawClass, 10)
	return VM{}
}

func (vm *VM) LoadClass(path string) {
	bytes, _ := ioutil.ReadFile(path)
	class, _ := ParseClass(bytes)
	vm.classes = append(vm.classes, class)
}

func (vm *VM) Run() {
	var stack []byte
	vm.classes[0].Execute("main", stack)
}

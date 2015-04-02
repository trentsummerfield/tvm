package java

import (
	"io/ioutil"
)

type VM struct {
	classes []class
}

func NewVM() (vm VM) {
	vm.classes = make([]class, 10)
	return VM{}
}

func (vm *VM) LoadClass(path string) {
	bytes, _ := ioutil.ReadFile(path)
	class, _ := parseClass(bytes)
	vm.classes = append(vm.classes, class)
}

func (vm *VM) Run() {
	var stack []byte
	vm.classes[0].execute("main", stack)
}

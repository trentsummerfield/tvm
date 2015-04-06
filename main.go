package main

import (
	"flag"

	"github.com/trentsummerfield/tvm/java"
)

func main() {
	flag.Parse()
	vm := java.NewVM()
	for _, class := range flag.Args() {
		vm.LoadClass(class)
	}
	vm.Run()
}

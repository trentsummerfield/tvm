package main

import (
	"flag"

	"github.com/trentsummerfield/tvm/java"
)

func main() {
	flag.Parse()
	filename := flag.Arg(0)

	vm := java.NewVM()
	vm.LoadClass(filename)
	vm.Run()
}

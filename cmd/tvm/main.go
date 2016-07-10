package main

import (
	"os"

	"github.com/trentsummerfield/tvm"
)

func main() {
	vm := java.NewVM()
	for _, arg := range os.Args[1:] {
		if isDirectory(arg) {
			vm.AddDirectory(arg)
		} else {
			vm.LoadClass(arg)
		}
	}
	vm.Run()
}

func isDirectory(arg string) bool {
	f, err := os.Open(arg)
	defer f.Close()
	if err != nil {
		return false
	}
	stats, err := f.Stat()
	if err != nil {
		return false
	}
	return stats.IsDir()
}

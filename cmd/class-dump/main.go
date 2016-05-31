package main

import (
	"fmt"
	"log"
	"os"

	tvm "github.com/trentsummerfield/tvm"
)

func main() {
	path := os.Args[1]
	file, err := os.Open(path)
	if err != nil {
		log.Panicf("unable to load %s: %v", path, err)
	}
	class, err := tvm.ParseClass(file)
	if err != nil {
		log.Panicf("unable to parse %s: %v", file, err)
	}

	fmt.Printf("Classfile %s\n", path)
	fmt.Printf("  Last modified %s; size %d bytes\n", "01/01/1970", 0)
	fmt.Printf("  MD5 checksum %s\n", "A_HASH")
	fmt.Printf("  Compiled from \"%s\"\n", "A java file :/")
	fmt.Printf("...\n")
	fmt.Printf("  minor version: %d\n", class.MinorVersion)
	fmt.Printf("  major version: %d\n", class.MajorVersion)
	fmt.Printf("  flags: %d\n", class.AccessFlags)

	fmt.Printf("Constant pool:\n")
	for i, cp := range class.ConstantPoolItems {
		_, ok := cp.(tvm.LongConstantPart2)
		if ok {
			continue
		}
		index := fmt.Sprintf("#%d", i+1)
		name := cp.String()
		fmt.Printf("  %4s = %s\n", index, name)
	}
}

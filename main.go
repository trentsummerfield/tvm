package main

import (
	"flag"
	"io/ioutil"

	"github.com/trentsummerfield/tvm/java"
)

func main() {
	flag.Parse()
	filename := flag.Arg(0)
	bytes, _ := ioutil.ReadFile(filename)
	class, _ := java.ParseClass(bytes)
	var stack []byte
	class.Execute("main", stack)
}

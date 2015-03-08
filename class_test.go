package main

import (
	"io/ioutil"
	"testing"
)

func TestFailure(t *testing.T) {
	bytes, err := ioutil.ReadFile("Hello.class")
	if err != nil {
		t.Fatalf("%v", err)
	}
	class, err := parse(bytes)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if class.magic != 0xCAFEBABE {
		t.Errorf("Magic number is wrong. Expected 0xCAFEBABE but got 0x%X", class.magic)
	}
	if class.minorVersion != 0 {
		t.Errorf("Minor number is wrong. Expected 0 but got %d", class.minorVersion)
	}
	if class.majorVersion != 52 {
		t.Errorf("Major number is wrong. Expected 52 but got %d", class.majorVersion)
	}
	if class.constantPoolCount != 29 {
		t.Errorf("Constant pool count is wrong. Expected 29 but got %d", class.constantPoolCount)
	}
}

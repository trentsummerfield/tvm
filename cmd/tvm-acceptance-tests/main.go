package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	tests, err := ioutil.ReadDir("tests")
	if err != nil {
		log.Fatal(err)
	}

	for _, test := range tests {
		if test.IsDir() {
			fmt.Println()
			fmt.Println(test.Name())
			files, err := filepath.Glob(filepath.Join("tests", test.Name(), "*.java"))
			if err != nil {
				log.Fatal("unable to find java files", err)
			}
			expected, err := ioutil.ReadFile(filepath.Join("tests", test.Name(), "out"))
			if err != nil {
				log.Fatal("unable to load expected output", err)
			}
			dir, err := ioutil.TempDir("", test.Name())
			if err != nil {
				log.Fatal("unable to create temp directory", err)
			}

			var javaOpts []string
			javaOpts = append(javaOpts, "-bootclasspath", "stdlib", "-d", dir)
			javaOpts = append(javaOpts, files...)
			javac := exec.Command("javac", javaOpts...)
			out, err := javac.CombinedOutput()
			if err != nil {
				log.Fatal("Failed to compile", string(out), err)
			}

			files, err = filepath.Glob(filepath.Join(dir, "*.class"))
			if err != nil {
				log.Fatal(err)
			}
			var tvmOpts []string
			tvmOpts = append(tvmOpts, "stdlib")
			tvmOpts = append(tvmOpts, files...)
			tvm := exec.Command("tvm", tvmOpts...)
			var tvmOut bytes.Buffer
			tvm.Stdout = &tvmOut
			tvm.Stderr = os.Stderr
			err = tvm.Run()
			if err != nil {
				fmt.Println("FAILED")
				log.Fatal(err)
			}

			expectedOut := removeCarriageReturns(string(expected))
			actualOut := removeCarriageReturns(tvmOut.String())

			if expectedOut != actualOut {
				fmt.Println("FAILED")
				fmt.Print("Expected:\n", expectedOut)
				fmt.Print("Actual:\n", actualOut)
				log.Fatal("")
			} else {
				fmt.Println("PASSED")
			}
		}
	}
}

func removeCarriageReturns(s string) string {
	return strings.Replace(s, "\r", "", -1)
}

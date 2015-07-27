// Written in 2015 by Jack Galilee. Convenient rights reserved.
// Use of this source code is governed by the MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	. "github.com/jgalilee/fractran/lang"
)

// Usage instructions for the program if fed through STDIN.
var PipeArgErr error = errors.New("USAGE: n")

// Usage instructions for the program if not fed through STDIN.
var FileArgErr error = errors.New("USAGE: file n")

// Check if program is being fed through STDIN.
func isPiped() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// Check the validity of the given arguments.
func checkArgs() error {
	nargs := len(os.Args)
	piped := isPiped()
	switch {
	case (2 != nargs) && piped:
		return PipeArgErr
	case (3 != nargs) && !piped:
		return FileArgErr
	}
	return nil
}

// Find the source of the program either from STDIN or the name of the file.
func findSrc() (io.Reader, error) {
	if isPiped() {
		return os.Stdin, nil
	}
	return os.Open(os.Args[1])
}

// Find the initial value of n to give the program.
func findN() (int64, error) {
	idx := 2
	if isPiped() {
		idx = 1
	}
	return strconv.ParseInt(os.Args[idx], 10, 64)
}

func parse(r io.Reader) (*Program, error) {
	parser := new(Parser)
	instrs, err := parser.Parse(r)
	if nil != err {
		return nil, err
	}
	return NewProgram(instrs)
}

func main() {
	var (
		n   int64
		src io.Reader
		err error
		prg *Program
	)
	// Check the arguments.
	if err = checkArgs(); nil != err {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	// Find the input source (file or stdin).
	src, err = findSrc()
	if nil != err {
		fmt.Fprintf(os.Stderr, "input error: %v\n", err)
	}
	// Find the initial value of n.
	n, err = findN()
	if nil != err {
		fmt.Fprintf(os.Stderr, "arg error: %v\n", err)
	}
	// Parse the source into a program.
	prg, err = parse(src)
	if nil != err {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}
	// Run the program until it halts.
	prg.Run(os.Stdout, n)
	os.Exit(0)
}

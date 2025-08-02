// Written in 2015 by Jack Galilee. Convenient rights reserved.
// Use of this source code is governed by the MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/jgalilee/fractran"
)

func main() {
	args := os.Args[1:]
	stat, _ := os.Stdin.Stat()
	isPiped := (stat.Mode() & os.ModeCharDevice) == 0

	var src io.Reader = os.Stdin
	var valueArg string

	switch {
	case isPiped && len(args) == 1:
		valueArg = args[0]
	case !isPiped && len(args) == 2:
		f, err := os.Open(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "fractran: open file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		src = f
		valueArg = args[1]
	default:
		fmt.Fprintln(os.Stderr, "Usage: fractran [file] n")
		os.Exit(1)
	}

	n, err := strconv.ParseInt(valueArg, 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fractran: parse number: %v\n", err)
		os.Exit(1)
	}

	instructions, err := new(fractran.Parser).Parse(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fractran: parse program: %v\n", err)
		os.Exit(1)
	}

	program, err := fractran.NewProgram(instructions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fractran: create program: %v\n", err)
		os.Exit(1)
	}

	if err := program.Run(os.Stdout, n); err != nil {
		fmt.Fprintf(os.Stderr, "fractran: run program: %v\n", err)
		os.Exit(1)
	}
}

# FRACTRAN

[![GoDoc](https://godoc.org/github.com/jgalilee/fractran?status.svg)](https://godoc.org/github.com/jgalilee/fractran)

Implementation of John Conway's esoteric FRACTRAN programming language in Go.

## What is FRACTRAN?

FRACTRAN is a Turing-complete esoteric programming language invented by mathematician John Conway. A FRACTRAN program consists of:

- A single positive integer (the initial state)
- A finite list of positive fractions (the instructions)

The program executes by repeatedly multiplying the current integer by the first fraction in the list that produces an integer result, until no such fraction exists.

Despite its simplicity, FRACTRAN can perform complex computations - Conway famously created a 14-instruction program that generates prime numbers!

## Installation

```bash
go install github.com/jgalilee/fractran/cmd/fractran@latest
```

## Quick Start

Here's a simple FRACTRAN program that halves numbers:

```bash
echo "1/2" | fractran 8
```

Output:
```
8
4
2
1
```

The program terminates when `1 × 1/2 = 1/2` is not an integer.

## Conway's Prime Generator

The most famous FRACTRAN program generates prime numbers. Create a file `primes.ft`:

```
17/91,78/85,19/51,23/38,29/33,77/29,95/23,77/19,1/17,11/13,13/11,15/14,15/2,55/1
```

Run it with initial value 2:

```bash
fractran primes.ft 2
```

Output:
```
2
15
825
725
1925
2275
425
390
330
290
770
```

This sequence encodes prime numbers! The powers of 2 in the prime factorisation of certain terms correspond to prime numbers (2, 3, 5, 7, 11, ...).

## CLI Usage

FRACTRAN supports both file input and piped input:

```bash
# From file
fractran primes.ft 2

# From pipe  
cat primes.ft | fractran 2

# Both commands are equivalent
```

The grammar currently disallows whitespace or newlines in the program source.

## Library Usage

Use FRACTRAN in your Go programs:

```go
package main

import (
    "fmt"
    "os"
    "strings"
    
    "github.com/jgalilee/fractran"
)

func main() {
    // Parse a FRACTRAN program
    parser := new(fractran.Parser)
    instrs, err := parser.Parse(strings.NewReader("3/2,1/3"))
    if err != nil {
        panic(err)
    }
    
    // Create program with step limit
    prog, err := fractran.NewBoundProgram(instrs, 10)
    if err != nil {
        panic(err)
    }
    
    // Run with initial value 4
    prog.Run(os.Stdout, 4)
}
```

The `github.com/jgalilee/fractran` package provides the `Program` and `Parser` types for programmatic use.

## Grammar

FRACTRAN programs use a simple grammar:

```
program    := fraction (',' fraction)*
fraction   := integer '/' integer
integer    := digit+
```

- Fractions must be positive
- No whitespace or newlines allowed
- Fractions separated by commas

## License

MIT License - see [LICENSE](LICENSE) file for details.
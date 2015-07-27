// Use of this source code is governed by the MIT-style
// license that can be found in the LICENSE file.
package lang

import (
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
)

// Halt is returned when for a given program either the maximum bound is
// reached, or nf is not a member of N+ for all f in L.
var Halt error = errors.New("fractran: done.")

// FRACTRAN program. It is made up of a starting integer n and a ordered list of
// positive fractions (instructions).
type Program struct {
	// Index of the last instruction executed, defaults to -1.
	Last int
	// Number of steps taken by the program.
	Steps int64
	// Maximum number of steps that can be taken (1,2,3,...,+Inf).
	Bound float64

	// Current value of n
	n      *big.Rat
	instrs []*big.Rat
}

// Returns a new FRACTRAN Program with a defined bound.
func NewBoundProgram(instrs []*big.Rat, b float64) (*Program, error) {
	// Validate the bound is only a positive integer.
	if math.IsInf(b, -1) && (b > 0) {
		return nil, fmt.Errorf("bound can't be %v", b)
	}
	// Validate the instructions are all positive fractions.
	for _, i := range instrs {
		if i.Sign() <= 0 {
			return nil, fmt.Errorf("%v is a non-positive fraction", i)
		}
	}
	return &Program{Last: -1, Bound: b, instrs: instrs}, nil
}

// Returns a new FRACTRAN program with an infinite maximum bound.
func NewProgram(instrs []*big.Rat) (*Program, error) {
	return NewBoundProgram(instrs, math.Inf(1))
}

// Steps the program as defined by the rules of FRACTRAN:
//
// 1) Find f in L where nf is an integer, assign nf to n.
// 2) Repeat 1 while there for at least one f in L nf is an integer.
//
// If the program was given a bound the number of steps for the programs is
// restricted to execute rule one a fixed number of times.
//
// The function will return Halt once the bound is reached or no instance of
// rule one holds.
func (p *Program) Step() (*big.Int, error) {
	if 0 == p.Steps {
		p.Steps++
		return p.n.Num(), nil
	}
	for j, i := range p.instrs {
		tmp := new(big.Rat)
		tmp.Mul(p.n, i)
		if tmp.IsInt() {
			p.Last = j
			p.n.Set(tmp)
			p.Steps++
			return p.n.Num(), nil
		}
	}
	return nil, Halt
}

// Continues to step the FRACTRAN program and writes the result to the output
// writer.
func (p *Program) Run(out io.Writer, n int64) error {
	if n <= 0 {
		return fmt.Errorf("n must be positive")
	}
	p.n = big.NewRat(n, 1)
	for p.Bound > float64(p.Steps) {
		if n, err := p.Step(); Halt != err {
			io.WriteString(out, fmt.Sprintf("%v\n", n))
		} else {
			return err
		}
	}
	return nil
}

// Continues to step the FRACTRAN program and writes the result to the output
// writer. In addition it list the instructions and highlights the instruction
// that gave the new value of n by enclosing it in square brackets.
func (p *Program) Debug(out io.Writer) error {
	for p.Bound > float64(p.Steps) {
		if n, err := p.Step(); Halt != err {
			io.WriteString(out, fmt.Sprintf("%v:", n))
			for j, i := range p.instrs {
				if p.Last == j {
					io.WriteString(out, fmt.Sprintf("\t[%v]", i))
				} else {
					io.WriteString(out, fmt.Sprintf("\t %v ", i))
				}
			}
		}
		io.WriteString(out, "\n")
	}
	return nil
}

// Written in 2015 by Jack Galilee. Convenient rights reserved.
// Use of this source code is governed by the MIT-style
// license that can be found in the LICENSE file.

// Package lang implements Conway's FRACTRAN programming language.
// FRACTRAN is a Turing-complete esoteric programming language that consists
// of a single positive integer and a finite list of positive fractions.
package fractran

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"strconv"
	"unicode"
)

// symbol represents a token in the syntax of the FRACTRAN programming language.
type symbol uint8

const (
	comma symbol = iota
	digit
	slash
	done
)

// String returns a string representation of the symbol.
func (s symbol) String() string {
	switch s {
	case comma:
		return "comma"
	case digit:
		return "digit"
	case slash:
		return "slash"
	case done:
		return "done"
	}
	return "unknown"
}

// Parser implements a recursive descent parser for FRACTRAN programs.
type Parser struct {
	currentSym  symbol
	currentRune rune
	lastRune    rune
	r           *bufio.Reader
}

// next reads the next rune from the input source and tokenizes it.
// If there was an error reading it panics. It assigns the current rune
// as the last rune before reading.
func (p *Parser) next() {
	var err error
	p.lastRune = p.currentRune
	p.currentRune, _, err = p.r.ReadRune()
	if err == io.EOF {
		p.currentSym = done
		return
	}
	if err != nil {
		panic(err)
	}
	// tokenize
	switch {
	case p.currentRune == ',':
		p.currentSym = comma
	case p.currentRune == '/':
		p.currentSym = slash
	case unicode.IsDigit(p.currentRune):
		p.currentSym = digit
	default:
		p.currentSym = done
	}
}

// current returns the current symbol.
func (p *Parser) current() symbol {
	return p.currentSym
}

// expect asserts that the current symbol is equal to the given symbol.
// If it is not, an error is returned.
func (p *Parser) expect(sym symbol) error {
	if sym != p.current() {
		return fmt.Errorf("unexpected symbol %v, expected %v", p.current(), sym)
	}
	return nil
}

// accept asserts that the current symbol is equal to the given symbol.
// It returns true if they are equal and false otherwise. If true, it
// advances to the next symbol.
func (p *Parser) accept(sym symbol) bool {
	if sym == p.current() {
		p.next()
		return true
	}
	return false
}

// integer reads the next integer from the input source.
func (p *Parser) integer() (int64, error) {
	var buff bytes.Buffer
	if err := p.expect(digit); err != nil {
		return -1, err
	}
	for p.accept(digit) {
		buff.WriteRune(p.lastRune)
	}
	return strconv.ParseInt(buff.String(), 10, 64)
}

// fraction reads the next fraction from the input source.
func (p *Parser) fraction() (*big.Rat, error) {
	var (
		num int64
		den int64
		err error
	)
	if num, err = p.integer(); err != nil {
		return nil, err
	}
	if err = p.expect(slash); err != nil {
		return nil, err
	}
	p.next()
	if den, err = p.integer(); err != nil {
		return nil, err
	}
	return big.NewRat(num, den), nil
}

// program reads the program from the input source.
func (p *Parser) program() ([]*big.Rat, error) {
	var (
		instr  *big.Rat
		instrs []*big.Rat
		err    error
	)
	p.next()
	if p.current() == done {
		return nil, fmt.Errorf("empty program")
	}
	for p.current() != done {
		instr, err = p.fraction()
		if err != nil {
			return nil, err
		}
		instrs = append(instrs, instr)
		for p.accept(comma) {
			instr, err = p.fraction()
			if err != nil {
				return nil, err
			}
			instrs = append(instrs, instr)
		}
	}
	return instrs, nil
}

// Parse parses the input source and returns the program as a slice of fractions.
func (p *Parser) Parse(src io.Reader) ([]*big.Rat, error) {
	p.r = bufio.NewReader(src)
	return p.program()
}

// Halt is returned when a FRACTRAN program terminates either because the maximum
// bound is reached, or no fraction f in the instruction list L produces an integer
// when multiplied by the current value n.
var Halt = errors.New("fractran: done")

// Program represents a FRACTRAN program consisting of a starting integer n and
// an ordered list of positive fractions (instructions).
type Program struct {
	// Last is the index of the last instruction executed, defaults to -1.
	Last int
	// Steps is the number of steps taken by the program.
	Steps int64
	// Bound is the maximum number of steps that can be taken (1,2,3,...,+Inf).
	Bound float64

	// Current value of n
	n      *big.Rat
	instrs []*big.Rat
}

// NewBoundProgram returns a new FRACTRAN Program with a defined step bound.
// The bound must be positive. All instruction fractions must be positive.
func NewBoundProgram(instrs []*big.Rat, b float64) (*Program, error) {
	// Validate the bound is positive
	if math.IsInf(b, -1) || b <= 0 {
		return nil, fmt.Errorf("bound must be positive")
	}
	// Validate the instructions are all positive fractions
	for _, i := range instrs {
		if i.Sign() <= 0 {
			return nil, fmt.Errorf("%v is a non-positive fraction", i)
		}
	}
	return &Program{Last: -1, Bound: b, instrs: instrs}, nil
}

// NewProgram returns a new FRACTRAN program with an infinite maximum bound.
func NewProgram(instrs []*big.Rat) (*Program, error) {
	return NewBoundProgram(instrs, math.Inf(1))
}

// Step executes one step of the FRACTRAN program as defined by the rules:
//  1. Find the first fraction f in L where n*f is an integer, assign n*f to n.
//  2. Repeat step 1 whilst there exists at least one f in L where n*f is an integer.
//
// Where L is the ordered list of fraction instructions.
//
// If the program was given a bound, the number of steps is restricted.
// The function returns Halt once the bound is reached or no fraction produces an integer.
func (p *Program) Step() (*big.Int, error) {
	if p.Steps == 0 {
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

// Run continues to step the FRACTRAN program and writes each result to the output writer.
// The initial value n must be positive.
func (p *Program) Run(out io.Writer, n int64) error {
	if n <= 0 {
		return fmt.Errorf("n must be positive")
	}
	p.n = big.NewRat(n, 1)
	for p.Bound > float64(p.Steps) {
		result, err := p.Step()
		if err == Halt {
			return nil // Normal termination
		}
		if err != nil {
			return err // Other error
		}
		io.WriteString(out, fmt.Sprintf("%v\n", result))
	}
	return nil
}

// Debug continues to step the FRACTRAN program and writes each result to the output writer.
// In addition, it lists the instructions and highlights the instruction that produced
// the new value of n by enclosing it in square brackets.
func (p *Program) Debug(out io.Writer) error {
	for p.Bound > float64(p.Steps) {
		result, err := p.Step()
		if err == Halt {
			break
		}
		if err != nil {
			return err
		}
		io.WriteString(out, fmt.Sprintf("%v:", result))
		for j, i := range p.instrs {
			if p.Last == j {
				io.WriteString(out, fmt.Sprintf("\t[%v]", i))
			} else {
				io.WriteString(out, fmt.Sprintf("\t%v", i)) // Removed trailing space
			}
		}
		io.WriteString(out, "\n")
	}
	return nil
}

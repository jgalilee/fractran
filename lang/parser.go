// Written in 2015 by Jack Galilee. Convenient rights reserved.
// Use of this source code is governed by the MIT-style
// license that can be found in the LICENSE file.
package lang

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"unicode"
)

// Represents a symbol in the syntax of the FRACTRAN programming language.
type symbol uint8

const (
	comma symbol = iota
	digit
	slash
	done
)

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

// Implementation of a recursive decent parser for FRACTRAN programs.
type Parser struct {
	currentSym  symbol
	currentRune rune
	lastRune    rune
	r           *bufio.Reader
}

// Reads the next rune from the input source. If there was an error reading it
// panics. It assigns the current rune as the last rune before reading.
func (p *Parser) next() {
	var err error
	p.lastRune = p.currentRune
	p.currentRune, _, err = p.r.ReadRune()
	if io.EOF == err {
		p.currentSym = done
		return
	}
	if err != nil {
		panic(err)
	}
	// tokenize
	switch {
	// comma
	case p.currentRune == ',':
		p.currentSym = comma
	// slash
	case p.currentRune == '/':
		p.currentSym = slash
	// digit
	case unicode.IsDigit(p.currentRune):
		p.currentSym = digit
	// done
	default:
		p.currentSym = done
	}
}

// Returns the current symbol.
func (p *Parser) current() symbol {
	return p.currentSym
}

// Asserts that the current symbol is equal to the given symbol. If it is not
// an error is returned.
func (p *Parser) expect(sym symbol) error {
	if sym != p.current() {
		return fmt.Errorf("unexpected symbol %v, expected %v", p.current(), sym)
	}
	return nil
}

// Asserts that the current symbol is equal to the given symbol. It returns true
// if they are equal and false otherwise.
func (p *Parser) accept(sym symbol) bool {
	if sym == p.current() {
		p.next()
		return true
	}
	return false
}

// Read the next integer from the input source.
func (p *Parser) integer() (int64, error) {
	var buff bytes.Buffer
	if err := p.expect(digit); nil != err {
		return -1, err
	}
	for p.accept(digit) {
		buff.WriteRune(p.lastRune)
	}
	return strconv.ParseInt(buff.String(), 10, 64)
}

// Read the next fraction from the input source.
func (p *Parser) fraction() (*big.Rat, error) {
	var (
		num int64
		den int64
		err error
	)
	if num, err = p.integer(); nil != err {
		return nil, err
	}
	if err = p.expect(slash); nil != err {
		return nil, err
	}
	p.next()
	if den, err = p.integer(); nil != err {
		return nil, err
	}
	return big.NewRat(num, den), nil
}

// Read the program from the input source
func (p *Parser) program() ([]*big.Rat, error) {
	var (
		instr  *big.Rat
		instrs []*big.Rat
		err    error
	)
	p.next()
	for done != p.current() {
		instr, err = p.fraction()
		if nil != err {
			return nil, err
		}
		instrs = append(instrs, instr)
		for p.accept(comma) {
			instr, err = p.fraction()
			if nil != err {
				return nil, err
			}
			instrs = append(instrs, instr)
		}
	}
	return instrs, nil
}

// Parses the input source and returns the program
func (p *Parser) Parse(src io.Reader) ([]*big.Rat, error) {
	p.r = bufio.NewReader(src)
	return p.program()
}

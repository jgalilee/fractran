// Written in 2015 by Jack Galilee. Convenient rights reserved.
// Use of this source code is governed by the MIT-style
// license that can be found in the LICENSE file.
package lang

import (
	"bytes"
	"math/big"
	"os"
	"testing"
)

var src string = "17/91,78/85,19/51,23/38,29/33,77/29,95/23,77/19,1/17,11/13,13/11,15/14,15/2,55/1"

func TestParse(t *testing.T) {
	input := bytes.NewBufferString(src)
	parser := new(Parser)
	givenInstrs, err2 := parser.Parse(input)
	if nil != err2 {
		panic(err2)
	}
	expect := 14
	given := len(givenInstrs)
	if expect != given {
		t.Errorf("%v given, expected %v\n", given, expect)
		t.Fail()
	}
	expectInstrs := []*big.Rat{
		big.NewRat(17, 91),
		big.NewRat(78, 85),
		big.NewRat(19, 51),
		big.NewRat(23, 38),
		big.NewRat(29, 33),
		big.NewRat(77, 29),
		big.NewRat(95, 23),
		big.NewRat(77, 19),
		big.NewRat(1, 17),
		big.NewRat(11, 13),
		big.NewRat(13, 11),
		big.NewRat(15, 14),
		big.NewRat(15, 2),
		big.NewRat(55, 1),
	}
	for i := range givenInstrs {
		givenInstr := givenInstrs[i]
		expectInstr := expectInstrs[i]
		if 0 != givenInstr.Cmp(expectInstr) {
			t.Errorf("%v given, expected %v", givenInstr, expectInstr)
			t.Fail()
		}
	}
}

func ExampleRunFromSource() {
	input := bytes.NewBufferString(src)
	parser := new(Parser)
	instrs, err1 := parser.Parse(input)
	if nil != err1 {
		panic(err1)
	}
	prg, err2 := NewBoundProgram(instrs, 11)
	if nil != err2 {
		panic(err2)
	}
	prg.Run(os.Stdout, 2)
	// Output:
	// 2
	// 15
	// 825
	// 725
	// 1925
	// 2275
	// 425
	// 390
	// 330
	// 290
	// 770
}

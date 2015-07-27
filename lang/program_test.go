// Written in 2015 by Jack Galilee. Convenient rights reserved.
// Use of this source code is governed by the MIT-style
// license that can be found in the LICENSE file.
package lang

import (
	"math/big"
	"os"
)

func ExampleRun() {
	inst := []*big.Rat{
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
	prg, err := NewBoundProgram(inst, 11)
	if nil != err {
		panic(err)
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

// Written in 2015 by Jack Galilee. Convenient rights reserved.
// Use of this source code is governed by the MIT-style
// license that can be found in the LICENSE file.
package fractran

import (
	"bytes"
	"math"
	"math/big"
	"os"
	"strings"
	"testing"
)

// Conway's prime-generating FRACTRAN program
const primeProgram = "17/91,78/85,19/51,23/38,29/33,77/29,95/23,77/19,1/17,11/13,13/11,15/14,15/2,55/1"

// Simple test programs
const simpleProgram = "3/2,1/3"      // Multiply by 3/2, then by 1/3 = divide by 2
const singleFraction = "2/1"         // Double the input
const terminatingProgram = "1/2,1/3" // Halve, then third (terminates on odd numbers)

func TestParser(t *testing.T) {
	t.Run("ValidPrograms", func(t *testing.T) {
		tests := []struct {
			name   string
			input  string
			expect int
		}{
			{"prime program", primeProgram, 14},
			{"simple program", simpleProgram, 2},
			{"single fraction", singleFraction, 1},
			{"terminating program", terminatingProgram, 2},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				parser := new(Parser)
				instrs, err := parser.Parse(strings.NewReader(tt.input))
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(instrs) != tt.expect {
					t.Errorf("got %d instructions, expected %d", len(instrs), tt.expect)
				}
			})
		}
	})

	t.Run("PrimeInstructionValues", func(t *testing.T) {
		parser := new(Parser)
		instrs, err := parser.Parse(strings.NewReader(primeProgram))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := []*big.Rat{
			big.NewRat(17, 91), big.NewRat(78, 85), big.NewRat(19, 51), big.NewRat(23, 38),
			big.NewRat(29, 33), big.NewRat(77, 29), big.NewRat(95, 23), big.NewRat(77, 19),
			big.NewRat(1, 17), big.NewRat(11, 13), big.NewRat(13, 11), big.NewRat(15, 14),
			big.NewRat(15, 2), big.NewRat(55, 1),
		}

		for i, instr := range instrs {
			if i >= len(expected) {
				t.Errorf("got more instructions than expected")
				break
			}
			if instr.Cmp(expected[i]) != 0 {
				t.Errorf("instruction %d: got %v, expected %v", i, instr, expected[i])
			}
		}
	})

	t.Run("InvalidPrograms", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"empty input", ""},
			{"invalid fraction", "1//2"},
			{"missing denominator", "1/"},
			{"missing numerator", "/2"},
			{"non-numeric", "a/b"},
			{"incomplete", "1/2,"},
			{"extra comma", "1/2,,3/4"},
			{"no slash", "12"},
			{"trailing slash", "1/2/"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				parser := new(Parser)
				_, err := parser.Parse(strings.NewReader(tt.input))
				if err == nil {
					t.Errorf("expected error for input %q", tt.input)
				}
			})
		}
	})
}

func TestProgramCreation(t *testing.T) {
	t.Run("ValidPrograms", func(t *testing.T) {
		tests := []struct {
			name    string
			instrs  []*big.Rat
			bound   float64
			wantErr bool
		}{
			{"positive fractions", []*big.Rat{big.NewRat(1, 2), big.NewRat(3, 4)}, 100, false},
			{"single instruction", []*big.Rat{big.NewRat(2, 1)}, 10, false},
			{"infinite bound", []*big.Rat{big.NewRat(1, 2)}, math.Inf(1), false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var prog *Program
				var err error

				if tt.bound > 0 {
					prog, err = NewBoundProgram(tt.instrs, tt.bound)
				}

				if tt.wantErr {
					if err == nil {
						t.Error("expected error but got none")
					}
					return
				}

				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if prog == nil {
					t.Fatal("program is nil")
				}
			})
		}
	})

	t.Run("InvalidPrograms", func(t *testing.T) {
		tests := []struct {
			name   string
			instrs []*big.Rat
			bound  float64
		}{
			{"negative fraction", []*big.Rat{big.NewRat(-1, 2)}, 10},
			{"zero fraction", []*big.Rat{big.NewRat(0, 1)}, 10},
			{"negative bound", []*big.Rat{big.NewRat(1, 2)}, -1},
			{"zero bound", []*big.Rat{big.NewRat(1, 2)}, 0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := NewBoundProgram(tt.instrs, tt.bound)
				if err == nil {
					t.Error("expected error but got none")
				}
			})
		}
	})
}

func TestProgramExecution(t *testing.T) {
	t.Run("StepByStep", func(t *testing.T) {
		// Simple program: 3/2 (multiply by 1.5)
		instrs := []*big.Rat{big.NewRat(3, 2)}
		prog, err := NewBoundProgram(instrs, 5)
		if err != nil {
			t.Fatalf("failed to create program: %v", err)
		}

		// Start with n=2
		prog.n = big.NewRat(2, 1)

		// First step should return initial value
		result, err := prog.Step()
		if err != nil {
			t.Fatalf("step 0 failed: %v", err)
		}
		if result.String() != "2" {
			t.Errorf("step 0: got %s, expected 2", result)
		}

		// Second step: 2 * 3/2 = 3
		result, err = prog.Step()
		if err != nil {
			t.Fatalf("step 1 failed: %v", err)
		}
		if result.String() != "3" {
			t.Errorf("step 1: got %s, expected 3", result)
		}
	})

	t.Run("Termination", func(t *testing.T) {
		// Program that only works on even numbers: 1/2
		instrs := []*big.Rat{big.NewRat(1, 2)}
		prog, err := NewBoundProgram(instrs, 10)
		if err != nil {
			t.Fatalf("failed to create program: %v", err)
		}

		// Start with n=4, should get: 4 -> 2 -> 1 -> halt
		var buf bytes.Buffer
		err = prog.Run(&buf, 4)
		if err != nil {
			t.Fatalf("run failed: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		expected := "4\n2\n1"
		if output != expected {
			t.Errorf("got output:\n%s\nexpected:\n%s", output, expected)
		}
	})

	t.Run("BoundLimit", func(t *testing.T) {
		// Infinite loop program: 2/1 (double forever)
		instrs := []*big.Rat{big.NewRat(2, 1)}
		prog, err := NewBoundProgram(instrs, 3) // Limit to 3 steps
		if err != nil {
			t.Fatalf("failed to create program: %v", err)
		}

		var buf bytes.Buffer
		err = prog.Run(&buf, 1)
		if err != nil {
			t.Fatalf("run failed: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Errorf("got %d output lines, expected 3", len(lines))
		}
	})

	t.Run("InvalidInput", func(t *testing.T) {
		instrs := []*big.Rat{big.NewRat(1, 2)}
		prog, err := NewProgram(instrs)
		if err != nil {
			t.Fatalf("failed to create program: %v", err)
		}

		var buf bytes.Buffer
		err = prog.Run(&buf, -1) // Negative input
		if err == nil {
			t.Error("expected error for negative input")
		}

		err = prog.Run(&buf, 0) // Zero input
		if err == nil {
			t.Error("expected error for zero input")
		}
	})
}

func TestDebugOutput(t *testing.T) {
	// Program where first instruction can't be used
	instrs := []*big.Rat{big.NewRat(1, 3), big.NewRat(3, 2)} // 1/3 (only works on multiples of 3), 3/2
	prog, err := NewBoundProgram(instrs, 3)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}

	// Start with n=2: 2 * 1/3 = 2/3 (not integer), 2 * 3/2 = 3 (integer)
	prog.n = big.NewRat(2, 1)

	var buf bytes.Buffer
	err = prog.Debug(&buf)
	if err != nil {
		t.Fatalf("debug failed: %v", err)
	}

	output := buf.String()
	// Should show which instruction was used (highlighted with brackets)
	if !strings.Contains(output, "[3/2]") {
		t.Error("debug output should highlight the used instruction")
	}
	if !strings.Contains(output, "1/3") {
		t.Error("debug output should show unused instructions")
	}
}

func TestHaltCondition(t *testing.T) {
	instrs := []*big.Rat{big.NewRat(1, 2)} // Only works on even numbers
	prog, err := NewProgram(instrs)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}

	prog.n = big.NewRat(3, 1) // Start with odd number

	// First step returns the initial value
	result, err := prog.Step()
	if err != nil {
		t.Fatalf("first step failed: %v", err)
	}
	if result.String() != "3" {
		t.Errorf("first step: got %s, expected 3", result)
	}

	// Second step should halt (3 * 1/2 = 3/2, not an integer)
	_, err = prog.Step()
	if err != Halt {
		t.Errorf("expected Halt error, got %v", err)
	}
}

// Benchmark parsing performance
func BenchmarkParsePrimeProgram(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parser := new(Parser)
		_, err := parser.Parse(strings.NewReader(primeProgram))
		if err != nil {
			b.Fatalf("parse failed: %v", err)
		}
	}
}

// Benchmark execution performance
func BenchmarkRunSimpleProgram(b *testing.B) {
	instrs := []*big.Rat{big.NewRat(1, 2)}
	prog, err := NewBoundProgram(instrs, 10)
	if err != nil {
		b.Fatalf("failed to create program: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		prog.Steps = 0             // Reset for each iteration
		err = prog.Run(&buf, 1024) // Start with power of 2
		if err != nil {
			b.Fatalf("run failed: %v", err)
		}
	}
}

// Example demonstrating basic FRACTRAN usage
func ExampleProgram_Run() {
	// Simple program that halves numbers
	instrs := []*big.Rat{big.NewRat(1, 2)}
	prog, err := NewBoundProgram(instrs, 5)
	if err != nil {
		panic(err)
	}

	prog.Run(os.Stdout, 8)
	// Output:
	// 8
	// 4
	// 2
	// 1
}

// Example demonstrating debug output
func ExampleProgram_Debug() {
	// Program with two instructions
	instrs := []*big.Rat{big.NewRat(3, 2), big.NewRat(1, 2)}
	prog, err := NewBoundProgram(instrs, 3)
	if err != nil {
		panic(err)
	}

	prog.n = big.NewRat(2, 1)
	prog.Debug(os.Stdout)
	// Output:
	// 2:	3/2	1/2
	// 3:	[3/2]	1/2
}

// Example showing Conway's prime generator
func ExampleParser_Parse() {
	parser := new(Parser)
	instrs, err := parser.Parse(strings.NewReader(primeProgram))
	if err != nil {
		panic(err)
	}

	// Run Conway's prime-generating program
	prog, err := NewBoundProgram(instrs, 11)
	if err != nil {
		panic(err)
	}

	prog.Run(os.Stdout, 2)
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

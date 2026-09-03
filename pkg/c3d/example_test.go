package c3d_test

import (
	"fmt"

	"gitlab.com/mireu-lab/Load-motion-animation/pkg/c3d"
)

func ExampleNewParser() {
	parser := c3d.NewParser()
	fmt.Printf("Parser initialized: %t\n", parser != nil)
	// Output: Parser initialized: true
}

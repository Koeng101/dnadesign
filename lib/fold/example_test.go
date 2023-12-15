package fold_test

import (
	"fmt"

	"github.com/koeng101/dnadesign/lib/fold"
)

func ExampleZuker() {
	result, _ := fold.Zuker("ACCCCCUCCUUCCUUGGAUCAAGGGGCUCAA", 37.0)
	brackets := result.DotBracket()
	fmt.Println(brackets)
	// Output: .((((.(((......)))....))))
}

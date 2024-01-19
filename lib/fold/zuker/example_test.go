package zuker_test

import (
	"fmt"

	zuker "github.com/koeng101/dnadesign/lib/fold/zuker"
)

func ExampleZuker() {
	result, _ := zuker.Zuker("ACCCCCUCCUUCCUUGGAUCAAGGGGCUCAA", 37.0)
	brackets := result.DotBracket()
	fmt.Println(brackets)
	// Output: .((((.(((......)))....))))
}

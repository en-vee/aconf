package aconf

import (
	"fmt"
	"testing"
)

const fileContents = `
*
name = "axlrate-imdg"
?
axlrate { # Main block
	name = "axlrate-imdg"
	# Another comment
	// * # This is an invalid character
	imdg {
		timeout = 10 seconds # number of seconds
		name = "axlrate-imdg"
	}
}	
`

func TestTokenize(t *testing.T) {
	l := HoconLexer{InputString: fileContents}

	//items := l.Run()
	items, errs := l.Run()
	for _, token := range items {
		fmt.Printf("%v\n", token)
	}

	fmt.Printf("\nErrors\n")
	for _, err := range errs {
		fmt.Printf("%v\n", err)
	}
}

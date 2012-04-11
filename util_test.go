package main

import "testing"

var yesShort = []string{
	"asdf",
	"as12df",
	"as_df",
}

var noShort = []string{
	"(!&#$(&@!#",
	"_asdf",
	"as   df",
	"ASDF",
	"AS_DF",
	"2o45",
	"as.12df",
}

func TestIsShortName(t *testing.T) {
	for _, y := range yesShort {
		if !isShortName(y) {
			t.Errorf("Is short: " + y)
		}
	}
	for _, n := range noShort {
		if isShortName(n) {
			t.Errorf("Is not short: " + n)
		}
	}
}

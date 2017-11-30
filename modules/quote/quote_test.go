// The MIT License (MIT)
//
// Copyright (c) 2017 Arnaud Vazard
//
// See LICENSE file.
package quote

import (
	"testing"
)

func Test_prepareForSearch(t *testing.T) {
	str := `This IS a; string.with, punctuation, quotes',		and upper: "case"  `
	expectedResult := "this is a string with punctuation quotes and upper case"
	result := prepareForSearch(str)
	if result != expectedResult {
		t.Errorf("Test result differ from expected result: \n Test result:\t%#v\nExpected result: %#v\n\n", result, expectedResult)
	}
}

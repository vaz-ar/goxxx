// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.
package memo

import (
	"fmt"
	"strings"
)

func HandleMemoCmd(message string, callback func(string)) bool {
	if strings.HasPrefix(message, "!memo") {
		fmt.Println("Found a memo")
		return true
	}
	return false
}

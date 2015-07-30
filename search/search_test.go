// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.
package search

import (
	"io/ioutil"
	"log"
	"testing"
)

var (
	terms          string = "Unit Tests"
	expectedResult string = "https://en.wikipedia.org/wiki/Unit_testing"
	mockFile       string = "./tests_data/duckduckgo.html"
)

func Test_getDuckduckgoPage(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	if getDuckduckgoPage(&terms) == nil {
		t.Errorf("No page returned for the search terms %q", terms)
	}
}

func Test_getDuckduckgoResult(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	content, err := ioutil.ReadFile(mockFile)
	if err != nil {
		t.Error("Failed to open the html file")
		t.FailNow()
	}

	if result := getDuckduckgoResult(content); result == nil {
		t.Error("No result returned by getDuckduckgoResult")
	} else if string(result) != expectedResult {
		t.Errorf("Expected result: %q, got %q instead\n", expectedResult, result)
	}
}

func Test_getDuckduckgoSearchResult(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	if result := getDuckduckgoSearchResult(&terms); result == nil {
		t.Error("No result returned by getDuckduckgoSearchResult")
	} else if string(result) != expectedResult {
		t.Errorf("Expected result: %q, got %q instead\n", expectedResult, result)
	}
}

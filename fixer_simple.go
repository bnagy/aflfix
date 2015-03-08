// +build simple

// To get tests to work properly you will need to `go build -tags simple`
// because the test code runs the aflfix server out of the current directory,
// which is not modified by the test invocation ( so it needs to be explicitly
// rebuilt including this fixer )

package main

import (
	"bytes"
)

type fixer struct{}

func NewFixer() Fixer {
	return &fixer{}
}

// change these!
var tests = map[string]string{
	"Blah Hello World": "Blah A MUCH LONGER THING World",
	"Blah":             "Blah",
	"Blah\xff\xfe\xaa\x00\x00Hello World": "Blah\xff\xfe\xaa\x00\x00A MUCH LONGER THING World",
	"": "",
}

func (f *fixer) Banner() string {
	return "Simple Example"
}

func (f *fixer) Fix(in []byte) ([]byte, error) {
	return bytes.Replace(in, []byte("Hello"), []byte("A MUCH LONGER THING"), -1), nil
}

func (f *fixer) BenchString() string {
	return "Blah\xff\xfe\xaa\x00\x00Hello World"
}

func (f *fixer) TestMap() map[string]string {
	return tests
}

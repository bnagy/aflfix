// your build tags would go here!
// eg: +build myfixer
// then go build -tags myfixer

package main

import (
	"bytes"
)

type fixer struct{}

func NewFixer() Fixer {
	return &fixer{}
}

func (f *fixer) Fix(in []byte) ([]byte, error) {
	return bytes.Replace(in, []byte("Hello"), []byte("A MUCH LONGER THING"), -1), nil
}

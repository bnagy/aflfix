// +build startxref

package main

import (
	"bytes"
	"fmt"
)

type fixer struct{}

func NewFixer() Fixer {
	return &fixer{}
}

var tests = map[string]string{
	pdfOld:          pdfNew,
	fragEmpty:       fragEmpty,
	fragNeither:     fragNeither,
	fragNoStartxref: fragNoStartxref,
	fragNoXref:      fragNoXref,
	doubleXrefs:     doubleXrefsFix,
}
var xref = []byte("xref")
var startxref = []byte("startxref")

// Grab some reuseable memory
var scratch = make([]byte, 0, 1024*1024*10)

func (f *fixer) Banner() string {
	return "Startxref 2.0"
}

func (f *fixer) Fix(in []byte) ([]byte, error) {

	sxrIdx := bytes.LastIndex(in, startxref)
	if sxrIdx < 0 {
		return in, nil
	}

	xrIdx := bytes.LastIndex(in[:sxrIdx], xref)
	if xrIdx < 0 {
		return in, nil
	}

	// This reslices scratch a lot, but keeps the same backing array
	scratch = scratch[:0]
	// These appends will grow the underlying array if required.
	scratch = append(scratch, in[:sxrIdx]...)
	scratch = append(scratch, []byte(fmt.Sprintf("startxref\n%d\n%%%%EOF\n", xrIdx))...)
	return scratch, nil
}

func (f *fixer) BenchString() string {
	return pdfNew
}

func (f *fixer) TestMap() map[string]string {
	return tests
}

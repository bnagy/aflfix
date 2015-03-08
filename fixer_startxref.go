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
}
var xref = []byte("xref")
var startxref = []byte("startxref")

// Grab some memory
var scratch = make([]byte, 0, 1024*1024*10)

func (f *fixer) Banner() string {
	return "Startxref 1.0"
}

func (f *fixer) Fix(in []byte) ([]byte, error) {
	xrIdx := bytes.Index(in, xref)
	if xrIdx < 0 {
		return in, nil
	}
	sxrIdx := bytes.Index(in, startxref)
	if sxrIdx < xrIdx {
		// either no xref token, so we got the idx of xref inside "startxref"
		// itself or no startxref so we got -1
		return in, nil
	}
	scratch = scratch[:0]
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

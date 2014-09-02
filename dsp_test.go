// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsp

import (
	"os"
	"testing"
)

func TestBasic(t *testing.T) {

	app := NewApp("Test", 1000)

	p1 := Source(4, 10).Use(NewNormal(88, 10, 2))

	w1 := app.Wire()
	app.Connect(p1, w1)

	p2 := WriteValues(os.Stdout, testing.Verbose())
	w2 := app.Wire()
	app.Connect(p2, w2, w1)

	for v := range w2 {
		_ = v
	}

	if app.Error() != nil {
		t.Fatalf("error: %s", app.Error())
	}
}

func TestWindow(t *testing.T) {

	app := NewApp("Test Window", 1000)

	p1 := Source(64, 2).Use(NewSquare(1, 0, 4, 4))

	w1 := app.Wire()
	app.Connect(p1, w1)

	p2 := Window(64).Use(Hamming)
	w2 := app.Wire()
	app.Connect(p2, w2, w1)

	p3 := WriteValues(os.Stdout, testing.Verbose())
	w3 := app.Wire()
	app.Connect(p3, w3, w2)

	// get a vector
	v := <-w3

	// check value
	hamming := HammingWindow(64)

	actual := v[0]
	expected := hamming[0]
	CompareFloats(t, expected, actual, "mismatched values in hamming window", 0.01)

	actual = v[4]
	expected = 0.0
	CompareFloats(t, expected, actual, "mismatched values in hamming window", 0.01)

	actual = v[57]
	expected = hamming[57]
	CompareFloats(t, expected, actual, "mismatched values in hamming window", 0.01)

	if app.Error() != nil {
		t.Fatalf("error: %s", app.Error())
	}
}

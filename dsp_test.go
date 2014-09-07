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
	app.ConnectOne(p1, w1)

	p2 := WriteValues(os.Stdout, testing.Verbose())
	w2 := app.Wire()
	app.ConnectOne(p2, w2, w1)

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
	app.ConnectOne(p1, w1)

	p2 := Window(64).Use(Hamming)
	w2 := app.Wire()
	app.ConnectOne(p2, w2, w1)

	p3 := WriteValues(os.Stdout, testing.Verbose())
	w3 := app.Wire()
	app.ConnectOne(p3, w3, w2)

	if app.Error() != nil {
		t.Fatalf("error: %s", app.Error())
	}

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
}

func TestChain(t *testing.T) {

	app := NewApp("Test Chain", 1000)

	out := app.Run(
		Source(64, 2).Use(NewSquare(1, 0, 4, 4)),
		Window(64).Use(Hamming),
		WriteValues(os.Stdout, testing.Verbose()),
	)

	if app.Error() != nil {
		t.Fatalf("error: %s", app.Error())
	}

	// get a vector
	v := <-out

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

}

func TestTwoWindows(t *testing.T) {

	app := NewApp("Test Two Windows", 1000)

	p1 := Source(64, 2).Use(NewSquare(1, 0, 4, 4))

	w1a := app.Wire()
	w1b := app.Wire()
	pio := PIO{Out: []ToChan{w1a, w1b}}
	app.Connect(p1, pio)

	p2a := Window(64).Use(Hamming)
	w2a := app.Wire()
	app.ConnectOne(p2a, w2a, w1a)

	p2b := Window(64).Use(Blackman)
	w2b := app.Wire()
	app.ConnectOne(p2b, w2b, w1b)

	p3a := WriteValues(os.Stdout, testing.Verbose())
	w3a := app.Wire()
	app.ConnectOne(p3a, w3a, w2a)

	p3b := WriteValues(os.Stdout, testing.Verbose())
	w3b := app.Wire()
	app.ConnectOne(p3b, w3b, w2b)

	if app.Error() != nil {
		t.Fatalf("error: %s", app.Error())
	}

	// get a vector from the Hamming window.
	v := <-w3a

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

	// get a vector from the Blackman window.
	v = <-w3b

	// check value
	blackman := BlackmanWindow(64)

	actual = v[0]
	expected = blackman[0]
	CompareFloats(t, expected, actual, "mismatched values in hamming window", 0.01)

	actual = v[4]
	expected = 0.0
	CompareFloats(t, expected, actual, "mismatched values in hamming window", 0.01)

	actual = v[57]
	expected = blackman[57]
	CompareFloats(t, expected, actual, "mismatched values in hamming window", 0.01)

}

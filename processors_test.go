// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsp

import (
	"os"
	"strings"
	"testing"
)

func TestAddScaled(t *testing.T) {

	const input = "0.0\t\t\n1 2 3.0 4\n5\t6.00 7\n\n8 9 10 11 12 13 14\n"
	r1 := strings.NewReader(input)
	r2 := strings.NewReader(input)

	app := NewApp("Test Reader", 10)

	p1 := Reader(r1, NewReader(4))
	p2 := Reader(r2, NewReader(4))
	p3 := AddScaled(4, 1.5)
	p4 := WriteValues(os.Stdout, testing.Verbose())

	w1 := app.Wire()
	w2 := app.Wire()
	w3 := app.Wire()
	w4 := app.Wire()

	pio1 := PIO{Out: []ToChan{w1}}
	app.Connect(p1, pio1)

	pio2 := PIO{Out: []ToChan{w2}}
	app.Connect(p2, pio2)

	pio3 := PIO{Out: []ToChan{w3}, In: []FromChan{w1, w2}}
	app.Connect(p3, pio3)

	app.ConnectOne(p4, w4, w3)

	if app.Error() != nil {
		t.Fatalf("error: %s", app.Error())
	}

	v := <-w4
	actual := v[3]
	expected := 9.0
	CompareFloats(t, expected, actual, "mismatched values", 0.01)

	v = <-w4
	actual = v[3]
	expected = 21.0
	CompareFloats(t, expected, actual, "mismatched values", 0.01)

}

func TestJoin(t *testing.T) {

	const input = "0.0\t\t\n1 2 3.0 4\n5\t6.00 7\n\n8 9 10 11 12 13 14\n"
	r1 := strings.NewReader(input)
	r2 := strings.NewReader(input)

	app := NewApp("Test Reader", 10)

	p1 := Reader(r1, NewReader(4))
	p2 := Reader(r2, NewReader(4))
	p3 := Join()
	p4 := WriteValues(os.Stdout, testing.Verbose())

	w1 := app.Wire()
	w2 := app.Wire()
	w3 := app.Wire()
	w4 := app.Wire()

	pio1 := PIO{Out: []ToChan{w1}}
	app.Connect(p1, pio1)

	pio2 := PIO{Out: []ToChan{w2}}
	app.Connect(p2, pio2)

	pio3 := PIO{Out: []ToChan{w3}, In: []FromChan{w1, w2}}
	app.Connect(p3, pio3)

	app.ConnectOne(p4, w4, w3)

	if app.Error() != nil {
		t.Fatalf("error: %s", app.Error())
	}

	v := <-w4
	actual := v[3]
	expected := 3.0
	CompareFloats(t, expected, actual, "mismatched values", 0.01)

	actualSize := len(v)
	expectedSize := 8
	if actualSize != expectedSize {
		t.Fatalf("mismatched length: %d vs. %d", actualSize, expectedSize)
	}

	actual = v[7]
	expected = 3
	CompareFloats(t, expected, actual, "mismatched values", 0.01)

	v = <-w4
	actual = v[6]
	expected = 6.0
	CompareFloats(t, expected, actual, "mismatched values", 0.01)

	actualSize = len(v)
	expectedSize = 8
	if actualSize != expectedSize {
		t.Fatalf("mismatched length: %d vs. %d", actualSize, expectedSize)
	}
}

func TestMovingAverage(t *testing.T) {

	const input = "1 3 5 3 1 3 13 -5 -3 -5"

	// expected output for winSize=4
	expected := []float64{1, 2, 3, 3, 3, 3, 5, 3, 2, 0}

	r := strings.NewReader(input)

	app := NewApp("Test MA", 10)

	out := app.Run(
		Reader(r, NewReader(1)),
		MovingAverage(1, 4, nil),
		WriteValues(os.Stdout, testing.Verbose()),
	)

	if app.Error() != nil {
		t.Fatalf("error: %s", app.Error())
	}

	for _, v := range expected {
		actual := <-out
		CompareFloats(t, v, actual[0], "mismatched values", 0.01)
	}
}

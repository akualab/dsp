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

func TestBasicReader(t *testing.T) {

	const input = "0.0\t\t\n1 2 3.0 4\n5\t6.00 7\n\n8 9 10 11 12 13 14\n"
	r := strings.NewReader(input)

	app := NewApp("Test Reader", 10)

	out := app.Run(
		Reader(r, NewReader(4)),
		WriteValues(os.Stdout, testing.Verbose()),
	)

	if app.Error() != nil {
		t.Fatalf("error: %s", app.Error())
	}

	//	for v := range out {
	//		_ = v
	//	}

	v := <-out
	t.Log(v)

}

func TestTextReader(t *testing.T) {

	const input = "\n\n\f1 2 3.0 4\n5\t6.00 7\n\n8 9 10 11 12 13 14 15 16 17 18\v19\r20\f\n\n\n21 22 23 24"
	r := strings.NewReader(input)

	app := NewApp("Test Reader", 10)

	c := &ReaderConfig{
		FrameSize: 5,
		StepSize:  2,
		ValueType: Text,
	}

	out := app.Run(
		Reader(r, c),
		WriteValues(os.Stdout, testing.Verbose()),
	)

	if app.Error() != nil {
		t.Fatalf("error: %s", app.Error())
	}

	//	for v := range out {
	//		_ = v
	//	}

	v := <-out
	actual := v[3]
	expected := 1.0
	CompareFloats(t, expected, actual, "mismatched values", 0.01)

	v = <-out
	actual = v[1]
	expected = 1.0
	CompareFloats(t, expected, actual, "mismatched values", 0.01)

	v = <-out
	actual = v[4]
	expected = 6.0
	CompareFloats(t, expected, actual, "mismatched values", 0.01)

	t.Log("chan len: ", len(out))
}

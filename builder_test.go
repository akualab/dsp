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

func TestBuilderSimple(t *testing.T) {

	const input = "0.0\t\t\n1 2 3.0 4\n5\t6.00 7\n\n8 9 10 11 12 13 14\n"
	r1 := strings.NewReader(input)

	app := NewApp("Test Builder Simple", 10)

	p1 := Reader(r1, NewReader(4))
	p2 := WriteValues(os.Stdout, testing.Verbose())

	b := app.NewBuilder()
	b.Add("reader 1", p1)
	b.Add("writer", p2)
	b.AddEndNode("end")

	var err error
	err = b.Connect("reader 1", "writer")
	CheckError(t, err)
	err = b.Connect("writer", "end")
	CheckError(t, err)

	b.Run()
	var ch chan Value
	ch, err = b.EndNodeChan("end")
	CheckError(t, err)
	_ = <-ch
}

func TestBuilder(t *testing.T) {

	const input = "0.0\t\t\n1 2 3.0 4\n5\t6.00 7\n\n8 9 10 11 12 13 14\n"
	r1 := strings.NewReader(input)
	r2 := strings.NewReader(input)

	app := NewApp("Test Builder", 10)

	p1 := Reader(r1, NewReader(4))
	p2 := Reader(r2, NewReader(4))
	p3 := AddScaled(4, 1.5)
	p4 := WriteValues(os.Stdout, testing.Verbose())

	b := app.NewBuilder()
	b.Add("reader 1", p1)
	b.Add("reader 2", p2)
	b.Add("combo", p3)
	b.Add("writer", p4)
	b.AddEndNode("end")

	var err error
	err = b.ConnectOrdered("reader 1", "combo", 1)
	CheckError(t, err)
	err = b.ConnectOrdered("reader 2", "combo", 0)
	CheckError(t, err)
	err = b.Connect("combo", "writer")
	CheckError(t, err)
	err = b.Connect("writer", "end")
	CheckError(t, err)

	b.Run()
	var ch chan Value
	ch, err = b.EndNodeChan("end")
	CheckError(t, err)
	v := <-ch

	actual := v[3]
	expected := 9.0
	CompareFloats(t, expected, actual, "mismatched values", 0.01)

	v = <-ch
	actual = v[3]
	expected = 21.0
	CompareFloats(t, expected, actual, "mismatched values", 0.01)

}

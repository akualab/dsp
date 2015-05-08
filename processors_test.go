// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsp

import (
	"fmt"
	"os"
	"testing"
)

func TestAddScaled(t *testing.T) {

	dim := 2
	app := NewApp("Test")

	app.Add("s1", Source(dim, 20, NewRandom(99)))
	app.Add("s2", Source(dim, 20, NewRandom(99)))
	app.Add("p1", AddScaled(dim, 0.5))
	app.Add("p2", WriteValues(os.Stdout, testing.Verbose()))

	app.Connect("p1", "s1", "s2")
	app.Connect("p2", "p1")

	s1 := app.NewTap("s1")
	s2 := app.NewTap("s2")
	out := app.NewTap("p2")
	var i int
	for ; i < 10; i++ {
		v1, _ := s1.Get(i)
		v2, _ := s2.Get(i)
		t.Log(v1, v2)

		exp := []float64{}
		for i, _ := range v1.Data {
			exp = append(exp, (v1.Data[i]+v2.Data[i])*0.5)
		}
		v, e := out.Get(i)
		if e != nil {
			t.Fatal(e)
		}
		CompareSliceFloat(t, exp, v.Data, "mismatch", 0.001)
		t.Log(v)
	}
}

func TestJoin(t *testing.T) {

	dim := 4
	app := NewApp("Test")

	app.Add("s1", Source(dim, 40, NewCounter()))
	app.Add("s2", Source(dim, 40, NewCounter()))
	app.Add("join", Join())

	app.Connect("join", "s1", "s2")
	out := app.NewTap("join")
	for k := 0; k < 2; k++ {
		var i int
		for ; i < 20; i++ {
			v, e := out.Get(i)
			t.Log(v)
			if e != nil {
				t.Fatal(e)
			}
			for j := 0; j < dim; j++ {
				if v.Data[j] != v.Data[j+dim] {
					t.Fatalf("mismatch j:%d, v1:%f, v2%f", j, v.Data[j], v.Data[j+dim])
				}
				if v.Data[j] != float64(int(i)*dim+j) {
					t.Fatalf("mismatch j:%d, v1:%f, v2%f", j, v.Data[j], float64(int(i)*dim+j))
				}
			}
		}
	}
}

func TestMovingAverage(t *testing.T) {

	input := []float64{1, 3, 5, 3, 1, 3, 13, -5, -3, -5}

	// expected output for winSize=4
	expected := []float64{1, 2, 3, 3, 3, 3, 5, 3, 2, 0}

	app := NewApp("Test MA")
	app.Add("source", Source(1, len(input), NewSlice(input)))
	app.Add("moving average", NewMAProc(1, 4, 20))

	app.Connect("moving average", "source")
	out := app.NewTap("moving average")

	app.Reset()
	var i int
	for ; i < int(len(input)); i++ {
		v, e := out.Get(i)
		if e != nil {
			t.Fatal(e)
		}
		t.Log(i, v)
		if v.Data[0] != expected[i] {
			t.Fatalf("expected %f, got %f", expected[i], v)
		}
	}
	app.Reset()
}

func TestDiff(t *testing.T) {

	input := []float64{1, 1, 7, 6, 5, 2, 2, 3, 4, 5, -1}

	// expected output for winSize=4
	expected := []float64{4, 4, 4, 1, -5, -3, -1, 3, -3}
	coeff := []float64{0, 1}
	app := NewApp("Test Diff")
	app.Add("source", Source(1, len(input), NewSlice(input)))
	app.Add("diff", NewDiffProc(1, 20, coeff))

	app.Connect("diff", "source")
	out := app.NewTap("diff")

	app.Reset()
	for i := 0; i < int(len(input)); i++ {
		t.Log("i:", i)
		v, e := out.Get(i)
		if e == ErrOOB {

			if i == len(input)-len(coeff) {
				t.Log("clean end")
				break
			} else {
				t.Fatal(fmt.Errorf("bad termination, expected i=%d, got i=%d", len(input)-len(coeff), i))
			}
		}
		if e != nil {
			t.Fatal(e)
		}
		t.Log(i, v)
		if v.Data[0] != expected[i] {
			t.Fatalf("expected %f, got %f", expected[i], v.Data[0])
		}
	}
	app.Reset()
}

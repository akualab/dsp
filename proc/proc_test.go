// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proc

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/akualab/dsp"
	narray "github.com/akualab/narray/na64"
)

var randSrc = rand.NewSource(99)

type TVal []float64

func (v TVal) Copy() dsp.Value {
	n := len(v)
	v2 := make(TVal, n, n)
	copy(v2, v)
	return v2
}

func source(r *rand.Rand, dim, length int) dsp.Processer {
	return dsp.NewProc(length, func(idx int, in ...dsp.Processer) (dsp.Value, error) {
		if idx < 0 || idx >= length {
			return nil, dsp.ErrOOB
		}
		return narray.Rand(r, dim), nil
	})
}

func slice(data []float64) dsp.Processer {
	return dsp.NewProc(len(data), func(idx int, in ...dsp.Processer) (dsp.Value, error) {
		length := len(data)
		if idx < 0 || idx >= length {
			return nil, dsp.ErrOOB
		}
		return narray.NewArray(data[idx:idx+1], 1), nil
	})
}

func TestAddScaled(t *testing.T) {

	dim := 2
	app := dsp.NewApp("Test")
	r := rand.New(randSrc)
	s1 := app.Add("s1", source(r, dim, 20))
	s2 := app.Add("s2", source(r, dim, 20))
	p1 := app.Add("p1", AddScaled(dim, 0.5))
	p2 := app.Add("p2", WriteValues(os.Stdout, testing.Verbose()))

	app.Connect(p1, s1, s2)
	app.Connect(p2, p1)

	out := app.NodeByName("p2")
	var i int
	for ; i < 10; i++ {
		v1, _ := s1.Get(i)
		na1 := v1.(*narray.NArray)
		v2, _ := s2.Get(i)
		na2 := v2.(*narray.NArray)
		t.Log(na1, na2)
		exp := []float64{}
		for i, _ := range na1.Data {
			exp = append(exp, (na1.Data[i]+na2.Data[i])*0.5)
		}
		v, e := out.Get(i)
		if e != nil {
			t.Fatal(e)
		}
		na := v.(*narray.NArray)
		compareSliceFloat(t, exp, na.Data, "mismatch", 0.001)
		t.Log(v)
	}
}

func TestJoin(t *testing.T) {

	r1 := rand.New(rand.NewSource(99))
	r2 := rand.New(rand.NewSource(99))
	dim := 4
	app := dsp.NewApp("Test")
	s1 := app.Add("s1", source(r1, dim, 20))
	s2 := app.Add("s2", source(r2, dim, 20))
	join := app.Add("join", Join())
	app.Connect(join, s1, s2)
	out := join
	for k := 0; k < 2; k++ {
		var i int
		for ; i < 20; i++ {
			val, e := out.Get(i)
			if e != nil {
				t.Fatal(e)
			}
			v := val.(*narray.NArray)
			t.Log(v)
			for j := 0; j < dim; j++ {
				if v.Data[j] != v.Data[j+dim] {
					t.Fatalf("mismatch j:%d, v1:%f, v2:%f", j, v.Data[j], v.Data[j+dim])
				}
				//				if v.Data[j] != float64(int(i)*dim+j) {
				//					t.Fatalf("mismatch j:%d, v1:%f, v2:%f", j, v.Data[j], float64(int(i)*dim+j))
				//				}
			}
		}
	}
}

func TestMovingAverage(t *testing.T) {

	input := []float64{1, 3, 5, 3, 1, 3, 13, -5, -3, -5}

	// expected output for winSize=4
	expected := []float64{1, 2, 3, 3, 3, 3, 5, 3, 2, 0}

	app := dsp.NewApp("Test MA")
	src := app.Add("source", slice(input))
	ma := app.Add("moving average", NewMAProc(1, 4, 20))

	app.Connect(ma, src)
	out := ma

	app.Reset()
	var i int
	for ; i < int(len(input)); i++ {
		val, e := out.Get(i)
		if e != nil {
			t.Fatal(e)
		}
		v := val.(*narray.NArray)
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
	app := dsp.NewApp("Test Diff")
	//	src := app.Add("source", Source(1, len(input), NewSlice(input)))
	src := app.Add("source", slice(input))
	diff := app.Add("diff", NewDiffProc(1, 20, coeff))

	app.Connect(diff, src)
	out := diff

	app.Reset()
	for i := 0; i < int(len(input)); i++ {
		t.Log("i:", i)
		v, e := out.Get(i)
		if e == dsp.ErrOOB {

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
		na := v.(*narray.NArray)
		if na.Data[0] != expected[i] {
			t.Fatalf("expected %f, got %f", expected[i], na.Data[0])
		}
	}
	app.Reset()
}

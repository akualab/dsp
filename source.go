// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsp

import (
	"math/rand"

	narray "github.com/akualab/narray/na64"
)

// NumberReader - number generators must implement this interface.
type NumberReader interface {
	Next() float64
}

// SourceProc is a processor that generates data.
type SourceProc struct {
	nr   NumberReader
	dim  int
	data [][]float64
}

// Source returns a data generator processor.
// Uses a Random number generator with values between
// 0 and 1 by default.
func Source(dim, len int, nr NumberReader) *SourceProc {
	data := make([][]float64, len, len)
	for i := range data {
		vec := make([]float64, dim, dim)
		data[i] = vec
		for j := range vec {
			vec[j] = nr.Next()
		}
	}
	return &SourceProc{
		nr:   nr,
		dim:  dim,
		data: data,
	}
}

func (s *SourceProc) SetInputs(in ...Processer) {}
func (s *SourceProc) Reset()                    {}

// Get implements the dsp.Processer interface.
func (s *SourceProc) Get(idx uint32) (Value, error) {
	if int(idx) > len(s.data)-1 {
		return nil, ErrOOB
	}
	return narray.NewArray(s.data[idx], s.dim), nil
}

// Slice of floats.
type Slice struct {
	data []float64
	idx  int
}

// NewSlice returns new Slice.
func NewSlice(s []float64) *Slice {

	return &Slice{data: s}
}

func (s *Slice) Next() float64 {
	v := s.data[s.idx]
	s.idx++
	return v
}

// Counter returns 0,1,2,...
type Counter struct {
	count int
}

// NewCounter returns new counter.
func NewCounter() *Counter {

	return &Counter{}
}

// Next float value.
func (c *Counter) Next() float64 {
	v := float64(c.count)
	c.count++
	return v
}

// Random returns pseudo-random numbers between 0 an 1
type Random struct {
	r *rand.Rand
}

// NewRandom returns random generator.
func NewRandom(seed int64) *Random {

	return &Random{
		r: rand.New(rand.NewSource(seed)),
	}
}

// Next float value.
func (random *Random) Next() float64 {
	return random.r.Float64()
}

// Normal returns random numbers generated with a Normal distribution.
type Normal struct {
	r        *rand.Rand
	mean, sd float64
}

// NewNormal returns a new Normal random generator.
func NewNormal(seed int64, mean, sd float64) *Normal {

	return &Normal{
		r:    rand.New(rand.NewSource(seed)),
		mean: mean,
		sd:   sd,
	}
}

// Next float value.
func (n *Normal) Next() float64 {
	return n.r.NormFloat64()*n.sd + n.mean
}

// Square is a square signal.
type Square struct {
	// Num samples with high values.
	HighDuration int
	// Num samples with low values.
	LowDuration int
	// High value.
	High float64
	// Low value.
	Low   float64
	state int
}

// NewSquare returns square generator.
func NewSquare(high, low float64, highDur, lowDur int) *Square {

	return &Square{
		High:         high,
		Low:          low,
		HighDuration: highDur,
		LowDuration:  lowDur,
	}
}

// Next returns the next float value.
func (s *Square) Next() float64 {

	v := s.Low
	size := s.HighDuration + s.LowDuration
	s.state = s.state % size

	if s.state < s.HighDuration {
		v = s.High
	}
	s.state++
	return v
}

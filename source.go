// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsp

import "math/rand"

// Number generators must implement this interface.
type NumberReader interface {
	Next() float64
}

// A processor to generate data.
type SourceProc struct {
	nr           NumberReader
	length, size int
}

// Returns a data generator processor.
// Uses a Random number generator with values between
// 0 and 1 by default.
func Source(size, length int) *SourceProc {
	return &SourceProc{
		nr:     NewRandom(99),
		length: length,
		size:   size,
	}
}

// Set the desire type of data generator.
func (s *SourceProc) Use(nr NumberReader) *SourceProc {
	s.nr = nr
	return s
}

// Implements the dsp.Processor interface.
func (s *SourceProc) RunProc(in []FromChan, out []ToChan) error {
	for i := 0; i < s.length; i++ {
		v := make(Value, s.size, s.size)
		for j := 0; j < s.size; j++ {
			v[j] = s.nr.Next()
		}
		SendValue(v, out)
	}
	return nil
}

// Returns pseudo-random numbers between 0 an 1
type Random struct {
	r *rand.Rand
}

// Returns random generator.
func NewRandom(seed int64) *Random {

	return &Random{
		r: rand.New(rand.NewSource(seed)),
	}
}

func (random *Random) Next() float64 {
	return random.r.Float64()
}

// Returns random numbers generated with a Normal distribution.
type Normal struct {
	r        *rand.Rand
	mean, sd float64
}

// Returns a new Normal random generator.
func NewNormal(seed int64, mean, sd float64) *Normal {

	return &Normal{
		r:    rand.New(rand.NewSource(seed)),
		mean: mean,
		sd:   sd,
	}
}

func (n *Normal) Next() float64 {
	return n.r.NormFloat64()*n.sd + n.mean
}

// Returns a suare signal.
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

// Returns square generator.
func NewSquare(high, low float64, highDur, lowDur int) *Square {

	return &Square{
		High:         high,
		Low:          low,
		HighDuration: highDur,
		LowDuration:  lowDur,
	}
}

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

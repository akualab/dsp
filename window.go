// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsp

import (
	"fmt"
	"math"
)

const (
	Rectangular = iota
	Hanning
	Hamming
	Blackman
)

type WindowProc struct {
	WinSize    int
	WindowType int
	data       []float64
	err        error
}

// Returns a Windowing processor with a rectangular shape.
func Window(winSize int) *WindowProc {
	return &WindowProc{
		WinSize:    winSize,
		WindowType: Rectangular,
	}
}

func (win *WindowProc) Use(windowType int) *WindowProc {

	win.WindowType = windowType
	switch windowType {

	case Rectangular:

	case Hanning:
		win.data = HanningWindow(win.WinSize)
	case Hamming:
		win.data = HammingWindow(win.WinSize)
	case Blackman:
		win.data = BlackmanWindow(win.WinSize)
	default:
		win.err = fmt.Errorf("Unknow window type: %d", windowType)
	}
	return win
}

// Implements the dsp.Processor interface.
func (win *WindowProc) RunProc(in In, out Out) error {

	if win.err != nil {
		return win.err
	}

	for in := range in.From[0] {

		inSize := len(in)
		if win.WinSize > inSize {
			return fmt.Errorf("window size [%d] is larger than input vector size [%d]", win.WinSize, inSize)
		}
		v := make(Value, win.WinSize, win.WinSize)
		if win.WindowType == Rectangular {
			copy(v, in)
		} else {
			// Multiply by data in slice.
			for i, _ := range win.data {
				v[i] = in[i] * win.data[i]
			}
		}
		//arg.Out <- v
		SendValue(v, out)
	}
	return nil
}

// Returns a Hanning window.
// w(t) = 0.5  – 0.5 * cos(2 pi t / T)
func HanningWindow(n int) []float64 {
	data := make([]float64, n, n)
	for i := 0; i < n; i++ {
		data[i] = 0.5 * (1.0 - math.Cos(2.0*math.Pi*float64(i)/float64(n)))
	}
	return data
}

// Returns a Hamming window.
// w(t) = 0.54  – 0.46 * cos(2 pi t / T)
func HammingWindow(n int) []float64 {
	data := make([]float64, n, n)
	for i := 0; i < n; i++ {
		data[i] = 0.54 - 0.46*math.Cos(2.0*math.Pi*float64(i)/float64(n))
	}
	return data
}

// Returns a Blackman window.
// w(t) = 0.42  – 0.5 * cos(2pt/T) + 0.08 * cos(4pt/T)
func BlackmanWindow(n int) []float64 {
	data := make([]float64, n, n)
	for i := 0; i < n; i++ {
		data[i] = 0.42 - 0.5*math.Cos(2.0*math.Pi*float64(i)/float64(n)) +
			0.08*math.Cos(4.0*math.Pi*float64(i)/float64(n))
	}
	return data
}

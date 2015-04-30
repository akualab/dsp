// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsp

import (
	"fmt"
	"math"

	narray "github.com/akualab/narray/na64"
)

const (
	// Rectangular window.
	Rectangular = iota
	// Hanning window.
	Hanning
	// Hamming window.
	Hamming
	// Blackman window.
	Blackman
)

// WindowProc is a window processor.
type WindowProc struct {
	WinSize    int
	WindowType int
	data       []float64
	err        error
	inputs     []Processer
}

// Window returns a window processor with a rectangular shape.
func Window(winSize int) *WindowProc {
	return &WindowProc{
		WinSize:    winSize,
		WindowType: Rectangular,
	}
}

// Use sets the window type.
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

// SetInputs for this processor.
func (win *WindowProc) SetInputs(in ...Processer) {
	win.inputs = in
}

// Reset WindowProc processor.
func (win *WindowProc) Reset() {}

// Get implements the dsp.Processer interface.
func (win *WindowProc) Get(idx uint32) (Value, error) {
	vec, err := win.inputs[0].Get(idx)
	if err != nil {
		return nil, err
	}
	//	inSize := len(vec)
	inSize := vec.Shape[0]
	if win.WinSize > inSize {
		return nil, fmt.Errorf("window size [%d] is larger than input vector size [%d]", win.WinSize, inSize)
	}
	v := narray.New(win.WinSize)
	if win.WindowType == Rectangular {
		copy(v.Data, vec.Data)
	} else {
		// Multiply by data in slice.
		for i, _ := range win.data {
			v.Data[i] = vec.Data[i] * win.data[i]
		}
	}
	return v, nil
}

// WindowSlice Returns a window as a slice of float64.
func WindowSlice(winType, winSize int) ([]float64, error) {

	switch winType {
	case Rectangular:
		s := make([]float64, winSize, winSize)
		for i := range s {
			s[i] = 1
		}
		return s, nil
	case Hanning:
		return HanningWindow(winSize), nil
	case Hamming:
		return HammingWindow(winSize), nil
	case Blackman:
		return BlackmanWindow(winSize), nil
	default:
		return nil, fmt.Errorf("Unknow window type: %d", winType)
	}
}

// HanningWindow returns a Hanning window.
// w(t) = 0.5  – 0.5 * cos(2 pi t / T)
func HanningWindow(n int) []float64 {
	data := make([]float64, n, n)
	for i := 0; i < n; i++ {
		data[i] = 0.5 * (1.0 - math.Cos(2.0*math.Pi*float64(i)/float64(n)))
	}
	return data
}

// HammingWindow returns a Hanning window.
// w(t) = 0.54  – 0.46 * cos(2 pi t / T)
func HammingWindow(n int) []float64 {
	data := make([]float64, n, n)
	for i := 0; i < n; i++ {
		data[i] = 0.54 - 0.46*math.Cos(2.0*math.Pi*float64(i)/float64(n))
	}
	return data
}

// BlackmanWindow returns a Blackman window.
// w(t) = 0.42  – 0.5 * cos(2pt/T) + 0.08 * cos(4pt/T)
func BlackmanWindow(n int) []float64 {
	data := make([]float64, n, n)
	for i := 0; i < n; i++ {
		data[i] = 0.42 - 0.5*math.Cos(2.0*math.Pi*float64(i)/float64(n)) +
			0.08*math.Cos(4.0*math.Pi*float64(i)/float64(n))
	}
	return data
}

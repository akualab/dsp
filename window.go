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
	StepSize   int
	WinSize    int
	WindowType int
	data       []float64
	err        error
	inputs     []Processer
	Centered   bool
	*Proc
}

// NewWindowProc returns a windowing processor.
// Input must return all source data on index zero.
func NewWindowProc(stepSize, winSize, windowType int, centered bool) *WindowProc {
	win := &WindowProc{
		StepSize:   stepSize,
		WinSize:    winSize,
		WindowType: windowType,
		Centered:   centered,
		Proc:       NewProc(defaultBufSize, nil),
	}

	win.WindowType = windowType
	switch windowType {

	case Rectangular:
		win.data = RectangularWindow(win.WinSize)
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

// Get implements the dsp.Processer interface.
func (win *WindowProc) Get(idx int) (Value, error) {
	if idx < 0 {
		return nil, ErrOOB
	}
	val, ok := win.GetCache(idx)
	if ok {
		return val, nil
	}
	vec, err := win.inputs[0].(Framer).Get(0)
	if err != nil {
		return nil, err
	}
	inSize := vec.Shape[0]
	ss := win.StepSize
	ws := win.WinSize
	if ws > inSize {
		return nil, fmt.Errorf("window size [%d] is larger than input vector [%d]", win.WinSize, inSize)
	}
	v := narray.New(win.WinSize)

	pr := int(idx) * ss
	pq := pr
	if win.Centered {
		ps := pr + ss/2
		pq = ps - ws/2
	}
	pu := pq + ws
	if pu > inSize {
		return nil, ErrOOB
	}
	var i int
	if pq < 0 {
		// Reflect waveform.
		for i = 0; i < (-pq); i++ {
			v.Data[i] = vec.Data[-pq-i] * win.data[i]
		}
	}
	for ; i < ws; i++ {
		v.Data[i] = vec.Data[i+pq] * win.data[i]
	}
	win.SetCache(idx, v)
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

// RectangularWindow returns a rectangular window.
// w(t) = 1.0
func RectangularWindow(n int) []float64 {
	data := make([]float64, n, n)
	for i := 0; i < n; i++ {
		data[i] = 1.0
	}
	return data
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

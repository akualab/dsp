package stream

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
	StepSize   int
	WindowType int
	data       []float64
	err        error
}

// Returns a Windowing processor with a rectangular shape.
func Window(winSize, stepSize int) *WindowProc {
	return &WindowProc{
		WinSize:    winSize,
		StepSize:   stepSize,
		WindowType: Rectangular,
	}
}

func (win *WindowProc) Use(windowType int) *WindowProc {

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

// Implements the stream.Processor interface.
func (win *WindowProc) RunProc(arg Arg) error {

	if win.err != nil {
		return win.err
	}

	if win.WindowType == Rectangular {

		return nil
	}

	// Multiply by data in slice.
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

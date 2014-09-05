// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsp

import "github.com/gonum/floats"

// Adds frames from all inputs and scales the added values.
// Blocks until all input frames are available.
// Will panic if all input frames do not have the same size.
func AddScaled(size int, alpha float64) Processor {
	return ProcFunc(func(arg Arg) error {

		numInputs := len(arg.In)
		for {
			v := make(Value, size, size)
			for i := 0; i < numInputs; i++ {
				w, ok := <-arg.In[i]
				if !ok {
					goto DONE
				}
				floats.Add(v, w)
			}
			floats.Scale(alpha, v)
			SendValue(v, arg)
		}
	DONE:
		return nil
	})
}

// Concatenate input frames. Size of output frame is sum of input frame sizes.
// Blocks until all input frames are available.
func Cat() Processor {
	return ProcFunc(func(arg Arg) error {

		numInputs := len(arg.In)
		for {
			v := Value{} // reset the output frame.
			for i := 0; i < numInputs; i++ {
				w, ok := <-arg.In[i]
				if !ok {
					goto DONE
				}
				v = append(v, w...)
			}
			SendValue(v, arg)
		}
	DONE:
		return nil
	})
}

// Computes the real FFT energy of the input frame.
// See dsp.RealFT and dsp.DFTEnergy for details.
// output frame size will be 2^logSize
// the real fft size computed is 2 * frame_size
func SpectralEnergy(logSize uint) Processor {
	fs := 1 << logSize // output frame size
	dftSize := 2 * fs
	return ProcFunc(func(arg Arg) error {

		for data := range arg.In[0] {
			dft := make(Value, dftSize, dftSize) // TODO: do not allocate every time. use slice pool?
			copy(dft, data)                      // zero padded
			RealFT(dft, dftSize, true)
			egy := DFTEnergy(dft)
			SendValue(egy, arg)
		}

		return nil
	})
}

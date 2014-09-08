// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsp

import (
	"fmt"
	"math"

	"github.com/gonum/floats"
)

// Adds frames from all inputs and scales the added values.
// Blocks until all input frames are available.
// Will panic if all input frames do not have the same size.
func AddScaled(size int, alpha float64) Processor {
	return ProcFunc(func(in In, out Out) error {

		numInputs := len(in.From)
		for {
			v := make(Value, size, size)
			for i := 0; i < numInputs; i++ {
				w, ok := <-in.From[i]
				if !ok {
					goto DONE
				}
				floats.Add(v, w)
			}
			floats.Scale(alpha, v)
			SendValue(v, out)
		}
	DONE:
		return nil
	})
}

// Stack multiple input frames into a single frame. Size of output frame is sum of input frame sizes.
// Blocks until all input frames are available.
func Join() Processor {
	return ProcFunc(func(in In, out Out) error {

		numInputs := len(in.From)
		for {
			v := Value{} // reset the output frame.
			for i := 0; i < numInputs; i++ {
				w, ok := <-in.From[i]
				if !ok {
					goto DONE
				}
				v = append(v, w...)
			}
			SendValue(v, out)
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
	return ProcFunc(func(in In, out Out) error {

		for data := range in.From[0] {
			dft := make(Value, dftSize, dftSize) // TODO: do not allocate every time. use slice pool?
			copy(dft, data)                      // zero padded
			RealFT(dft, dftSize, true)
			egy := DFTEnergy(dft)
			SendValue(egy, out)
		}

		return nil
	})
}

var (
	MelFilterbankIndices      = []int{10, 11, 14, 17, 20, 23, 27, 30, 33, 36, 40, 45, 50, 56, 62, 69, 76, 84}
	MelFilterbankCoefficients = [][]float64{
		[]float64{1.0, 1.0, 1.0, 1.0, 0.66, 0.33},
		[]float64{0.33, 0.66, 1.0, 1.0, 1.0, 1.0, 0.66, 0.33},
		[]float64{0.33, 0.66, 1.0, 1.0, 1.0, 1.0, 0.66, 0.33},
		[]float64{0.33, 0.66, 1.0, 1.0, 1.0, 1.0, 0.75, 0.5, 0.25},
		[]float64{0.33, 0.66, 1.0, 1.0, 1.0, 1.0, 1.0, 0.66, 0.33},
		[]float64{0.25, 0.5, 0.75, 1.0, 1.0, 1.0, 1.0, 0.66, 0.33},
		[]float64{0.33, 0.66, 1.0, 1.0, 1.0, 1.0, 0.66, 0.33},
		[]float64{0.33, 0.66, 1.0, 1.0, 1.0, 1.0, 0.75, 0.5, 0.25},
		[]float64{0.33, 0.66, 1.0, 1.0, 1.0, 1.0, 1.0, 0.8, 0.6, 0.4, 0.2},
		[]float64{0.25, 0.5, 0.75, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 0.8, 0.6, 0.4, 0.2},
		[]float64{0.2, 0.4, 0.6, 0.8, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 0.83, 0.66, 0.5, 0.33, 0.16},
		[]float64{0.2, 0.4, 0.6, 0.8, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 0.83, 0.66, 0.5, 0.33, 0.16},
		[]float64{0.16, 0.33, 0.5, 0.66, 0.83, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 0.85, 0.71, 0.57, 0.42, 0.28, 0.14},
		[]float64{0.16, 0.33, 0.5, 0.66, 0.83, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 0.85, 0.71, 0.57, 0.42, 0.28, 0.14},
		[]float64{0.14, 0.28, 0.42, 0.57, 0.71, 0.85, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 0.875, 0.75, 0.625, 0.5, 0.375, 0.25, 0.125},
		[]float64{0.142, 0.285, 0.428, 0.571, 0.714, 0.857, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 0.88, 0.77, 0.66, 0.55, 0.44, 0.33, 0.22, 0.11},
		[]float64{0.125, 0.25, 0.375, 0.5, 0.625, 0.75, 0.875, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 0.88, 0.77, 0.66, 0.55, 0.44, 0.33, 0.22, 0.11},
		[]float64{0.11, 0.22, 0.33, 0.44, 0.55, 0.66, 0.77, 0.88, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0},
	}
)

// Computes filterbank energies using the provided indices and coefficients.
func Filterbank(indices []int, coeff [][]float64) Processor {
	nf := len(indices) // num filterbanks

	return ProcFunc(func(in In, out Out) error {

		for input := range in.From[0] {
			fb := make(Value, nf, nf)
			for i := 0; i < nf; i++ {
				for k := 0; k < len(coeff[i]); k++ {
					fb[i] += coeff[i][k] * input[indices[i]+k]
				}
			}
			SendValue(fb, out)
		}
		return nil
	})
}

// Natual logarithm processor.
func Log() Processor {

	return ProcFunc(func(in In, out Out) error {

		for data := range in.From[0] {
			size := len(data)
			v := make(Value, size, size)
			for k, w := range data {
				v[k] = math.Log(w)
			}
			SendValue(v, out)
		}

		return nil
	})
}

// Sum returns the sum of the elements of the input frame.
func Sum() Processor {

	return ProcFunc(func(in In, out Out) error {

		for data := range in.From[0] {
			SendValue(Value{floats.Sum(data)}, out)
		}
		return nil
	})
}

/*
MaxNorm returns a norm value as follows:

  define: y[n] = norm[n-1] * alpha where alpha < 1
  define: norm(v) as sqrt(v . v) where "." is the dot product.

  max[n] = max(y[n], norm(x[n])
*/
func MaxNorm(alpha float64) Processor {

	return ProcFunc(func(in In, out Out) error {

		norm := 0.0
		for v := range in.From[0] {

			y := norm * alpha
			norm = math.Sqrt(floats.Dot(v, v))
			max := math.Max(y, norm)
			SendValue(Value{max}, out)
		}
		return nil
	})
}

// Discrete Cosine Transform
func DCT(inSize, outSize int) Processor {

	dct := GenerateDCT(outSize+1, inSize)
	return ProcFunc(func(in In, out Out) error {

		for input := range in.From[0] {
			size := len(input)
			if inSize != size {
				return fmt.Errorf("mismatch in size [%d] and input frame size [%d]", inSize, size)
			}
			v := make(Value, outSize, outSize)

			for i := 1; i <= outSize; i++ {
				for j := 0; j < inSize; j++ {
					v[i-1] += input[j] * dct[i][j]
				}
			}
			SendValue(v, out)
		}
		return nil
	})
}

/*
Compute moving average for the last M samples.

  for i >= M:
                  i
  AVG=G[i] = 1/M * sum X[j]
                 j=i-M

  for 0 < i < M
                  i
  AVG[i] = 1/i * sum X[j]
                 j=0

  Where AVG is the output vector and X is the input vector.

Will panic if output size is different from input size.
If param avg in not nil, it will be used as the intial avg
for i < M.
*/
func MovingAverage(outSize, winSize int, avg Value) Processor {

	return ProcFunc(func(in In, out Out) error {
		sum := make(Value, outSize, outSize)
		buf := make([]Value, winSize, winSize)
		var i uint32

		for input := range in.From[0] {
			v := movingSum(int(i%uint32(winSize)), buf, sum, input)
			if i >= uint32(winSize) {
				floats.Scale(1.0/float64(winSize), v)
			} else if len(avg) == 0 {
				floats.Scale(1.0/float64(i+1), v)
			} else {
				copy(v, avg)
			}
			SendValue(v, out)
			i++
		}
		return nil
	})
}

// Updates sum by subtracting oldest sample and adding newest.
func movingSum(p int, buf []Value, sum, data Value) Value {

	// Replace oldest value in buffer with newest value.
	old := buf[p]
	buf[p] = data

	// Subtract oldest, add newest.
	if old != nil {
		floats.Sub(sum, old)
	}
	floats.Add(sum, data)
	return sum.Copy()
}

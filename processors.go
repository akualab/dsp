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

// AddScaled adds frames from all inputs and scales the added values.
// Blocks until all input frames are available.
// Will panic if input frame sizes don't match.
func AddScaled(size int, alpha float64) Processor {
	return ProcFunc(func(in []FromChan, out []ToChan) error {

		numInputs := len(in)
		for {
			v := make(Value, size, size)
			for i := 0; i < numInputs; i++ {

				w, ok := ValueOK(in[i])
				if !ok {
					return nil
				}
				floats.Add(v, w)
			}
			floats.Scale(alpha, v)
			SendValue(v, out)
		}
	})
}

// Sub subtracts input[1] from input[0].
// Blocks until all input frames are available.
// Will panic if input frame sizes don't match.
func Sub() Processor {
	return ProcFunc(func(in []FromChan, out []ToChan) error {

		if len(in) != 2 {
			return fmt.Errorf("proc has %d inputs, expected 2", len(in))
		}
		for {
			a, aok := ValueOK(in[0])
			if !aok {
				return nil
			}
			b, bok := ValueOK(in[1])
			if !bok {
				return nil
			}
			v := make(Value, len(a), len(a))
			copy(v, a)
			floats.Sub(v, b)
			SendValue(v, out)
		}
	})
}

// Join stacks multiple input vectors into a single vector. Output vector size equals sum of input vector sizes.
// Blocks until all input vectors are available.
func Join() Processor {
	return ProcFunc(func(in []FromChan, out []ToChan) error {

		numInputs := len(in)
		for {
			v := Value{}
			for i := 0; i < numInputs; i++ {
				w, ok := ValueOK(in[i])
				if !ok {
					return nil
				}
				v = append(v, w...)
			}
			SendValue(v, out)
		}
	})
}

// SpectralEnergy computes the real FFT energy of the input frame.
// See dsp.RealFT and dsp.DFTEnergy for details.
// The size of the output vector is 2^logSize.
// The real fft size computed is 2 * frame_size.
func SpectralEnergy(logSize uint) Processor {
	fs := 1 << logSize // output frame size
	dftSize := 2 * fs
	return ProcFunc(func(in []FromChan, out []ToChan) error {
		for {
			data, ok := ValueOK(in[0])
			if !ok {
				return nil
			}
			dft := make(Value, dftSize, dftSize) // TODO: do not allocate every time. use slice pool?
			copy(dft, data)                      // zero padded
			RealFT(dft, dftSize, true)
			egy := DFTEnergy(dft)
			SendValue(egy, out)
		}
	})
}

var (
	// MelFilterbankIndices are the indices of the filters in the filterbank.
	MelFilterbankIndices = []int{10, 11, 14, 17, 20, 23, 27, 30, 33, 36, 40, 45, 50, 56, 62, 69, 76, 84}
	// MelFilterbankCoefficients is a hardcoded filterbank for the speech example.
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

// Filterbank computes filterbank energies using the provided indices and coefficients.
func Filterbank(indices []int, coeff [][]float64) Processor {
	nf := len(indices) // num filterbanks
	return ProcFunc(func(in []FromChan, out []ToChan) error {

		for {
			input, ok := ValueOK(in[0])
			if !ok {
				return nil
			}
			fb := make(Value, nf, nf)
			for i := 0; i < nf; i++ {
				for k := 0; k < len(coeff[i]); k++ {
					fb[i] += coeff[i][k] * input[indices[i]+k]
				}
			}
			SendValue(fb, out)
		}
	})
}

// Log returns the natural logarithm of the input.
func Log() Processor {

	return ProcFunc(func(in []FromChan, out []ToChan) error {

		for {
			data, ok := ValueOK(in[0])
			if !ok {
				return nil
			}
			size := len(data)
			v := make(Value, size, size)
			for k, w := range data {
				v[k] = math.Log(w)
			}
			SendValue(v, out)
		}
	})
}

// Sum returns the sum of the elements of the input frame.
func Sum() Processor {

	return ProcFunc(func(in []FromChan, out []ToChan) error {

		for {
			data, ok := ValueOK(in[0])
			if !ok {
				return nil
			}
			SendValue(Value{floats.Sum(data)}, out)
		}
	})
}

/*
MaxNorm returns a norm value as follows:

  define: y[n] = norm[n-1] * alpha where alpha < 1
  define: norm(v) as sqrt(v . v) where "." is the dot product.

  max[n] = max(y[n], norm(x[n])
*/
func MaxNorm(alpha float64) Processor {

	return ProcFunc(func(in []FromChan, out []ToChan) error {

		norm := 0.0
		for {
			v, ok := ValueOK(in[0])
			if !ok {
				return nil
			}
			y := norm * alpha
			norm = math.Sqrt(floats.Dot(v, v))
			max := math.Max(y, norm)
			SendValue(Value{max}, out)
		}
	})
}

// DCT returns the Discrete Cosine Transform of the input vector.
func DCT(inSize, outSize int) Processor {

	dct := GenerateDCT(outSize+1, inSize)
	return ProcFunc(func(in []FromChan, out []ToChan) error {
		for {
			input, ok := ValueOK(in[0])
			if !ok {
				return nil
			}

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
	})
}

/*
MovingAverage computes the average for the last M samples.

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
If param avg in not nil, it will be used as the initial avg
for i < M.
*/
func MovingAverage(outSize, winSize int, avg Value) Processor {

	return ProcFunc(func(in []FromChan, out []ToChan) error {
		sum := make(Value, outSize, outSize)
		buf := make([]Value, winSize, winSize)
		var i uint32
		for {
			input, ok := ValueOK(in[0])
			if !ok {
				return nil
			}

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

/*
DiffProc computes a weighted difference between samples as follows:

             delta-1
    diff[i] = sum c_j * { x[i+j+1] - x[i-j-1] }
              j=0

   delta-1
    sum {c_j} = 1.0
    j=0

Note that this operation in non-causal and will result in a delay
of delta samples.

Param "size" must match the size of the input vectors.
Param "coeff" is the slice of coefficients.
*/
type DiffProc struct {
	size    int
	bufSize int
	delta   int
	buf     []Value
	coeff   []float64
}

// NewDiffProc returns a new diff processor.
func NewDiffProc(size int, coeff []float64) *DiffProc {
	delta := len(coeff)
	bufSize := 2*delta + 1
	dp := &DiffProc{
		delta:   delta,
		bufSize: bufSize,
		buf:     make([]Value, bufSize, bufSize),
		size:    size,
		coeff:   coeff,
	}
	// Init circular buffer with zero values.
	for j := range dp.buf {
		dp.buf[j] = make(Value, size, size)
	}
	return dp
}

// Reset the DiffProc processor. Prepares the processor for a new stream.
func (dp *DiffProc) Reset() {
	for i, vec := range dp.buf {
		for j := range vec {
			dp.buf[j][i] = 0
		}
	}
}

// RunProc implements the dsp.Processor interface.
func (dp *DiffProc) RunProc(in []FromChan, out []ToChan) error {

	i := 0
	for {
		input, ok := ValueOK(in[0])
		if !ok {
			return nil
		}

		if len(input) != dp.size {
			return fmt.Errorf("input vector size %d does not match size %d", len(input), dp.size)
		}
		// Store the input vector in the buffer.
		// Should be safe to store a reference if we follow
		// the convention to treat input vectors as read only.
		dp.buf[i] = input
		v := make(Value, dp.size, dp.size)
		for j := 0; j < dp.delta; j++ {
			minus := Modulo(i-j-1, dp.bufSize)
			plus := Modulo(i+j+1, dp.bufSize)
			floats.AddScaled(v, -dp.coeff[j], dp.buf[minus])
			floats.AddScaled(v, dp.coeff[j], dp.buf[plus])
			// fmt.Printf("i:%4d, j:%d, in:%3.f, buf[%d]:%.f, buf[%d]:%.f, v:%3.f, coeff:%.f\n", i, j, input[0], plus, buf[plus][0], minus, buf[minus][0], v[0], coeff[j])
		}
		SendValue(v, out)
		i++
		i = i % dp.bufSize
	}
}

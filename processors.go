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

const defaultBufSize = 100

// Value is an multidimensional array that satisfies the framer interface.
type Value *narray.NArray

// AddScaled adds frames from all inputs and scales the added values.
// Will panic if input frame sizes don't match.
func AddScaled(size int, alpha float64) Processer {
	return NewProc(defaultBufSize, func(idx uint32, in ...Processer) (Value, error) {
		numInputs := len(in)
		v := narray.New(size)
		for i := 0; i < numInputs; i++ {
			vec, err := in[i].Get(idx)
			if err != nil {
				return nil, err
			}
			narray.Add(v, v, vec)
		}
		narray.Scale(v, v, alpha)
		return v, nil
	})
}

// Sub subtracts input[1] from input[0].
// Will panic if input frame sizes don't match.
func Sub() Processer {
	return NewProc(defaultBufSize, func(idx uint32, in ...Processer) (Value, error) {
		if len(in) != 2 {
			return nil, fmt.Errorf("proc has %d inputs, expected 2", len(in))
		}
		vec0, e0 := in[0].Get(idx)
		if e0 != nil {
			return nil, e0
		}
		vec1, e1 := in[1].Get(idx)
		if e1 != nil {
			return nil, e1
		}
		return narray.Sub(nil, vec0, vec1), nil
	})
}

// Join stacks multiple input vectors into a single vector. Output vector size equals sum of input vector sizes.
// Blocks until all input vectors are available.
func Join() Processer {
	return NewProc(defaultBufSize, func(idx uint32, in ...Processer) (Value, error) {
		numInputs := len(in)
		v := []float64{}
		for i := 0; i < numInputs; i++ {
			vec, err := in[i].Get(idx)
			if err != nil {
				return nil, err
			}
			v = append(v, vec.Data...)
		}
		na := narray.NewArray(v, len(v))
		return na, nil
	})
}

// SpectralEnergy computes the real FFT energy of the input frame.
// See dsp.RealFT and dsp.DFTEnergy for details.
// The size of the output vector is 2^logSize.
// The real fft size computed is 2 * frame_size.
func SpectralEnergy(logSize uint) Processer {
	fs := 1 << logSize // output frame size
	dftSize := 2 * fs
	return NewProc(defaultBufSize, func(idx uint32, in ...Processer) (Value, error) {
		dft := make([]float64, dftSize, dftSize) // TODO: do not allocate every time. use slice pool?
		vec, err := in[0].Get(idx)
		if err != nil {
			return nil, err
		}
		copy(dft, vec.Data) // zero padded
		RealFT(dft, dftSize, true)
		egy := DFTEnergy(dft)
		return narray.NewArray(egy, len(egy)), nil
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
func Filterbank(indices []int, coeff [][]float64) Processer {
	nf := len(indices) // num filterbanks
	return NewProc(defaultBufSize, func(idx uint32, in ...Processer) (Value, error) {
		fb := make([]float64, nf, nf)
		for i := 0; i < nf; i++ {
			for k := 0; k < len(coeff[i]); k++ {
				vec, err := in[0].Get(idx)
				if err != nil {
					return nil, err
				}
				fb[i] += coeff[i][k] * vec.Data[indices[i]+k]
			}
		}
		return narray.NewArray(fb, len(fb)), nil
	})
}

// Log returns the natural logarithm of the input.
func Log() Processer {
	return NewProc(defaultBufSize, func(idx uint32, in ...Processer) (Value, error) {
		vec, err := in[0].Get(idx)
		if err != nil {
			return nil, err
		}
		return narray.Log(nil, vec), nil
	})
}

// Sum returns the sum of the elements of the input frame.
func Sum() Processer {
	return NewProc(defaultBufSize, func(idx uint32, in ...Processer) (Value, error) {
		vec, err := in[0].Get(idx)
		if err != nil {
			return nil, err
		}
		sum := narray.New(1)
		v := (*narray.NArray)(vec)
		sum.Set(v.Sum(), 0)
		return sum, nil
	})
}

/*
MaxNorm returns a norm value as follows:

  define: y[n] = norm[n-1] * alpha where alpha < 1
  define: norm(v) as sqrt(v . v) where "." is the dot product.

  max[n] = max(y[n], norm(x[n])

The max value is computed in the range {0...idx}
*/
func MaxNorm(bufSize int, alpha float64) Processer {
	return NewProc(bufSize, func(idx uint32, in ...Processer) (Value, error) {
		max := 0.0
		norm := 0.0
		var i uint32
		for ; i <= idx; i++ {
			y := norm * alpha
			vec, err := in[0].Get(idx)
			if err != nil {
				return nil, err
			}
			norm = math.Sqrt(narray.Dot(vec, vec))
			max = math.Max(y, norm)
		}
		res := narray.New(1)
		res.Set(max, 0)
		return res, nil
	})
}

// DCT returns the Discrete Cosine Transform of the input vector.
func DCT(inSize, outSize int) Processer {

	dct := GenerateDCT(outSize+1, inSize)
	return NewProc(defaultBufSize, func(idx uint32, in ...Processer) (Value, error) {

		input, err := in[0].Get(idx)
		if err != nil {
			return nil, err
		}
		size := input.Shape[0]
		if inSize != size {
			return nil, fmt.Errorf("mismatch in size [%d] and input frame size [%d]", inSize, size)
		}
		v := make([]float64, outSize, outSize)
		for i := 1; i <= outSize; i++ {
			for j := 0; j < inSize; j++ {
				v[i-1] += input.Data[j] * dct[i][j]
			}
		}
		return narray.NewArray(v, len(v)), nil
	})
}

/*
MAProc computes the average for the last M samples.

  for i >= M:
                  i
  AVG[i] = 1/M * sum X[j]
                 j=i-M+1

  for 0 < i < M
                  i
  AVG[i] = 1/(i+1) * sum X[j]
                 j=0

  Where AVG is the output vector and X is the input vector.

Will panic if output size is different from input size.
If param avg in not nil, it will be used as the initial avg
for i < M.
*/
type MAProc struct {
	dim, bufSize int
	winSize      uint32
	*Proc
}

// NewMAProc creates a new MA processor.
func NewMAProc(dim, winSize, bufSize int) *MAProc {
	ma := &MAProc{
		dim:     dim,
		bufSize: bufSize,
		winSize: uint32(winSize),
		Proc:    NewProc(bufSize, nil),
	}
	return ma
}

// Get implements the dsp.Processer interface.
func (ma *MAProc) Get(idx uint32) (Value, error) {
	val, ok := ma.GetCache(idx)
	if ok {
		return val, nil
	}

	c := 1.0 / float64(ma.winSize)
	start := idx - ma.winSize + 1
	if idx < ma.winSize {
		c = 1.0 / float64(idx+1)
		start = 0
	}
	sum := narray.New(ma.dim)
	// TODO: no need to add every time, use a circular buffer.
	for j := start; j <= idx; j++ {
		v, e := ma.Input(0).Get(j)
		if e != nil {
			return nil, e
		}
		narray.Add(sum, sum, v)
	}
	narray.Scale(sum, sum, c)
	ma.SetCache(idx, sum)
	return sum, nil
}

/*
DiffProc computes a weighted difference between samples as follows:

    for delta < i < N-delta-1:

             delta-1
    diff[i] = sum c_j * { x[i+j+1] - x[i-j-1] }
              j=0

    where x is the input data stream, i is the frame index. and N
    is the number of frames. For other frame indices replace delta with:

    for i <= delta : delta' = i  AND  for i >= N-delta-1: delta' = N-1-i

Param "dim" must match the size of the input vectors.
Param "coeff" is the slice of coefficients.
*/
type DiffProc struct {
	dim       int
	delta     int
	buf       []Value
	coeff     []float64
	cacheSize int
	cache     *cache
	*Proc
}

// NewDiffProc returns a new diff processor.
func NewDiffProc(dim, bufSize int, coeff []float64) *DiffProc {
	delta := len(coeff)
	dp := &DiffProc{
		delta: delta,
		dim:   dim,
		coeff: coeff,
		Proc:  NewProc(bufSize, nil),
	}
	return dp
}

// Get implements the dsp.Processer interface.
func (dp *DiffProc) Get(idx uint32) (Value, error) {

	val, ok := dp.GetCache(idx)
	if ok {
		return val, nil
	}
	sum := narray.New(dp.dim)
	var j uint32
	for ; j < uint32(dp.delta); j++ {
		plus, ep := dp.Input(0).Get(idx + j + 1)
		if ep == ErrOOB {
			break
		}
		if ep != nil {
			return nil, ep
		}
		minus, em := dp.Input(0).Get(idx - j - 1)
		if em == ErrOOB {
			break
		}
		if em != nil {
			return nil, em
		}
		narray.AddScaled(sum, plus, dp.coeff[j])
		narray.AddScaled(sum, minus, -dp.coeff[j])
	}

	dp.SetCache(idx, sum)
	return sum, nil
}

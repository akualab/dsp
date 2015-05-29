// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proc

import (
	"fmt"
	"math"

	"github.com/akualab/dsp"
	narray "github.com/akualab/narray/na64"
)

const defaultBufSize = 1000

// Value is an multidimensional array that satisfies the framer interface.
type Value *narray.NArray

// Scale returns a scaled vector.
func Scale(alpha float64) dsp.Processer {
	return dsp.NewProc(defaultBufSize, func(idx int, in ...dsp.Processer) (dsp.Value, error) {
		vec, err := dsp.Processers(in).Get(idx)
		if err != nil {
			return nil, err
		}
		return narray.Scale(nil, vec.(*narray.NArray), alpha), nil
	})
}

// AddScaled adds frames from all inputs and scales the added values.
// Will panic if input frame sizes don't match.
func AddScaled(size int, alpha float64) dsp.Processer {
	return dsp.NewProc(defaultBufSize, func(idx int, in ...dsp.Processer) (dsp.Value, error) {
		numInputs := len(in)
		v := narray.New(size)
		for i := 0; i < numInputs; i++ {
			vec, err := in[i].(dsp.Framer).Get(idx)
			if err != nil {
				return nil, err
			}
			narray.Add(v, v, vec.(*narray.NArray))
		}
		narray.Scale(v, v, alpha)
		return v, nil
	})
}

// Sub subtracts in1 from in0. The inputs can be of type Framer of OneValuer.
// (The method uses reflection to get the type. For higher performance, implement a custom processor.)
// Will panic if input frame sizes don't match.
func Sub() dsp.Processer {
	return dsp.NewProc(defaultBufSize, func(idx int, in ...dsp.Processer) (dsp.Value, error) {
		if len(in) != 2 {
			return nil, fmt.Errorf("proc Sub needs 2 inputs got %d", len(in))
		}
		vec0, e0 := dsp.Get(in[0], idx)
		if e0 != nil {
			return nil, e0
		}
		vec1, e1 := dsp.Get(in[1], idx)
		if e1 != nil {
			return nil, e1
		}
		return narray.Sub(nil, vec0.(*narray.NArray), vec1.(*narray.NArray)), nil
	})
}

// Join stacks multiple input vectors into a single vector. Output vector size equals sum of input vector sizes.
// Blocks until all input vectors are available.
func Join() dsp.Processer {
	return dsp.NewProc(defaultBufSize, func(idx int, in ...dsp.Processer) (dsp.Value, error) {
		numInputs := len(in)
		framers, err := dsp.Processers(in).CheckInputs(numInputs)
		if err != nil {
			return nil, err
		}
		v := []float64{}
		for i := 0; i < numInputs; i++ {
			vec, err := framers[i].Get(idx)
			if err != nil {
				return nil, err
			}
			v = append(v, vec.(*narray.NArray).Data...)
		}
		na := narray.NewArray(v, len(v))
		return na, nil
	})
}

// SpectralEnergy computes the real FFT energy of the input frame.
// FFT size is 2^(logSize+1) and the size of the output vector is 2^logSize.
// See dsp.RealFT and dsp.DFTEnergy for details.
func SpectralEnergy(logSize int) dsp.Processer {
	fs := 1 << uint(logSize) // output frame size
	dftSize := 2 * fs
	return dsp.NewProc(defaultBufSize, func(idx int, in ...dsp.Processer) (dsp.Value, error) {
		dft := make([]float64, dftSize, dftSize) // TODO: do not allocate every time. use slice pool?
		vec, err := dsp.Processers(in).Get(idx)
		if err != nil {
			return nil, err
		}
		copy(dft, vec.(*narray.NArray).Data) // zero padded
		RealFT(dft, dftSize, true)
		egy := DFTEnergy(dft)
		return narray.NewArray(egy, len(egy)), nil
	})
}

// Filterbank computes filterbank energies using the provided indices and coefficients.
func Filterbank(indices []int, coeff [][]float64) dsp.Processer {
	nf := len(indices) // num filterbanks
	return dsp.NewProc(defaultBufSize, func(idx int, in ...dsp.Processer) (dsp.Value, error) {
		vec, err := dsp.Processers(in).Get(idx)
		if err != nil {
			return nil, err
		}
		fb := make([]float64, nf, nf)
		for i := 0; i < nf; i++ {
			for k := 0; k < len(coeff[i]); k++ {
				fb[i] += coeff[i][k] * vec.(*narray.NArray).Data[indices[i]+k]
			}
		}
		return narray.NewArray(fb, len(fb)), nil
	})
}

// Log returns the natural logarithm of the input.
func Log() dsp.Processer {
	return dsp.NewProc(defaultBufSize, func(idx int, in ...dsp.Processer) (dsp.Value, error) {
		vec, err := dsp.Processers(in).Get(idx)
		if err != nil {
			return nil, err
		}
		return narray.Log(nil, vec.(*narray.NArray)), nil
	})
}

// Sum returns the sum of the elements of the input frame.
func Sum() dsp.Processer {
	return dsp.NewProc(defaultBufSize, func(idx int, in ...dsp.Processer) (dsp.Value, error) {
		vec, err := dsp.Processers(in).Get(idx)
		if err != nil {
			return nil, err
		}
		sum := narray.New(1)
		v := vec.(*narray.NArray)
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
func MaxNorm(bufSize int, alpha float64) dsp.Processer {
	return dsp.NewProc(bufSize, func(idx int, in ...dsp.Processer) (dsp.Value, error) {
		max := 0.0
		norm := 0.0
		for i := 0; i <= idx; i++ {
			y := norm * alpha
			vec, err := dsp.Processers(in).Get(idx)
			if err != nil {
				return nil, err
			}
			na := vec.(*narray.NArray)
			norm = math.Sqrt(narray.Dot(na, na))
			max = math.Max(y, norm)
		}
		res := narray.New(1)
		res.Set(max, 0)
		return res, nil
	})
}

// DCT returns the Discrete Cosine Transform of the input vector.
func DCT(inSize, outSize int) dsp.Processer {

	dct := GenerateDCT(outSize+1, inSize)
	return dsp.NewProc(defaultBufSize, func(idx int, in ...dsp.Processer) (dsp.Value, error) {

		input, err := dsp.Processers(in).Get(idx)
		if err != nil {
			return nil, err
		}
		size := input.(*narray.NArray).Shape[0]
		if inSize != size {
			return nil, fmt.Errorf("mismatch in size [%d] and input frame size [%d]", inSize, size)
		}
		v := make([]float64, outSize, outSize)
		for i := 1; i <= outSize; i++ {
			for j := 0; j < inSize; j++ {
				v[i-1] += input.(*narray.NArray).Data[j] * dct[i][j]
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
	winSize      int
	*dsp.Proc
}

// NewMAProc creates a new MA processor.
func NewMAProc(dim, winSize, bufSize int) *MAProc {
	ma := &MAProc{
		dim:     dim,
		bufSize: bufSize,
		winSize: winSize,
		Proc:    dsp.NewProc(bufSize, nil),
	}
	return ma
}

// Get implements the dsp.dsp.Processer interface.
func (ma *MAProc) Get(idx int) (dsp.Value, error) {
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
		v, e := ma.Framer(0).Get(j)
		if e != nil {
			return nil, e
		}
		narray.Add(sum, sum, v.(*narray.NArray))
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
	buf       []dsp.Value
	coeff     []float64
	cacheSize int
	*dsp.Proc
}

// NewDiffProc returns a new diff processor.
func NewDiffProc(dim, bufSize int, coeff []float64) *DiffProc {
	delta := len(coeff)
	dp := &DiffProc{
		delta: delta,
		dim:   dim,
		coeff: coeff,
		Proc:  dsp.NewProc(bufSize, nil),
	}
	return dp
}

// Get implements the dsp.dsp.Processer interface.
func (dp *DiffProc) Get(idx int) (dsp.Value, error) {
	if idx < 0 {
		return nil, dsp.ErrOOB
	}
	val, ok := dp.GetCache(idx)
	if ok {
		return val, nil
	}
	res := narray.New(dp.dim)
	for j := 0; j < dp.delta; j++ {
		plus, ep := dp.Framer(0).Get(idx + j + 1)
		if ep != nil {
			return nil, ep
		}
		minus, em := dp.Framer(0).Get(idx - j - 1)
		if em == dsp.ErrOOB {
			// Repeat next frame.
			v, em := dp.Get(idx + 1)
			if em != nil {
				return nil, em
			}
			res = v.(*narray.NArray)
			break
		}
		if em != nil {
			return nil, em
		}
		narray.AddScaled(res, plus.(*narray.NArray), dp.coeff[j])
		narray.AddScaled(res, minus.(*narray.NArray), -dp.coeff[j])
	}
	dp.SetCache(idx, res)
	return res, nil
}

// MaxXCorrIndex returns the lag that maximizes the cross-correlation between two inputs.
// The param lagLimit is the highest lag value to be explored.
// Input vectors may have different lengths.
//  xcor[i] = x[n] * y[n-i]
// Returns the value of i that maximizes xcorr[i] and the max correlation value in a two-dimensional vector.
// value[0]=lag, value[1]=xcorr
func MaxXCorrIndex(lagLimit int) dsp.Processer {
	return dsp.NewProc(defaultBufSize, func(idx int, in ...dsp.Processer) (dsp.Value, error) {
		if len(in) != 2 {
			return nil, fmt.Errorf("proc Corr needs 2 inputs got %d", len(in))
		}
		if idx < 0 {
			return nil, fmt.Errorf("got negative index: %d", idx)
		}
		vec0, e0 := in[0].(dsp.Framer).Get(idx)
		if e0 != nil {
			return nil, e0
		}
		vec1, e1 := in[1].(dsp.Framer).Get(idx)
		if e1 != nil {
			return nil, e1
		}
		maxLag := 0
		maxCorr := -math.MaxFloat64
		n0 := len(vec0.(*narray.NArray).Data)
		n1 := len(vec1.(*narray.NArray).Data)
		for lag := 0; lag < lagLimit; lag++ {
			end := n0
			if n1+lag < end {
				end = len(vec1.(*narray.NArray).Data) + lag
			}
			if lag > end {
				break
			}
			sum := 0.0
			for i := lag; i < end; i++ {
				sum += vec0.(*narray.NArray).Data[i] * vec1.(*narray.NArray).Data[i-lag]
			}
			if sum > maxCorr {
				maxCorr = sum
				maxLag = lag
			}
		}
		return narray.NewArray([]float64{float64(maxLag), maxCorr}, 2), nil
	})
}

// MaxWin returns the elementwise max vector of the input stream.
func MaxWin() dsp.Processer {
	return dsp.NewOneProc(func(in ...dsp.Processer) (dsp.Value, error) {
		var max *narray.NArray
		var i int
		for {
			vec, err := dsp.Processers(in).Get(i)
			if err == dsp.ErrOOB {
				return max, nil
			}
			if err != nil {
				return nil, err
			}
			if i == 0 {
				max = narray.New(vec.(*narray.NArray).Shape[0])
				max.SetValue(-math.MaxFloat64)
			}
			narray.MaxArray(max, vec.(*narray.NArray), max)
			i++
		}
	})
}

// Mean returns the mean vector of the input stream.
//         N-1
//  mean = sum in_frame[i] where mean and in_frame are vectors.
//         i=0
func Mean() dsp.Processer {
	return dsp.NewOneProc(func(in ...dsp.Processer) (dsp.Value, error) {
		var mean *narray.NArray
		var i int
		for {
			vec, err := dsp.Processers(in).Get(i)
			if err == dsp.ErrOOB {
				return narray.Scale(mean, mean, 1/float64(i)), nil
			}
			if err != nil {
				return nil, err
			}
			if i == 0 {
				mean = narray.New(vec.(*narray.NArray).Shape[0])
			}
			narray.Add(mean, mean, vec.(*narray.NArray))
			i++
		}
	})
}

// MSE returns the mean squared error of two inputs.
func MSE() dsp.Processer {
	return dsp.NewProc(defaultBufSize, func(idx int, in ...dsp.Processer) (dsp.Value, error) {
		framers, err := dsp.Processers(in).CheckInputs(2)
		if err != nil {
			return nil, err
		}
		vec0, e0 := framers[0].Get(idx)
		if e0 != nil {
			return nil, e0
		}
		vec1, e1 := framers[1].Get(idx)
		if e1 != nil {
			return nil, e1
		}
		n := float64(vec0.(*narray.NArray).Shape[0])
		mse := narray.Sub(nil, vec0.(*narray.NArray), vec1.(*narray.NArray))
		narray.Mul(mse, mse, mse)
		narray.Scale(mse, mse, 1.0/n)
		return mse, nil
	})
}

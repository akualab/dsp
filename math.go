package dsp

import "math"

/*
Generate the Discrete Cosine Transform.

     for i = 0,..,N-1

              M-1
     dct[i] = sum x[j] * cos(i(2j+1)PI/M)
              j=0

     Return the following N x M transformation matrix:

     T(0,0)   T(0,1)   T(0,2)   ... T(0,M-1)
     T(1,0)   T(1,1)   T(1,2)   ... T(1,M-1)
     T(2,0)   T(2,1)   T(2,2)   ... T(2,M-1)
     ...
     T(N-1,0) T(N-1,1) T(N-1,2) ... T(N-1,M-1)

*/
func GenerateDCT(N, M int) [][]float64 {

	dct := make([][]float64, N, N)

	for i := 0; i < N; i++ {
		dct[i] = make([]float64, M, M)
		for j := 0; j < M; j++ {
			dct[i][j] = math.Cos(float64(i) * (2.0*float64(j) + 1.0) * math.Pi / float64(M))
		}
	}
	return dct
}

/*
Compute the complex FFT in-place.
( Adapted from Numerical Recipes Book)

data is the input array of length 2*nn.
nn the length of the discrete signal.
direct=true for direct FFT. Inverse otherwise.
*/
func four1(data []float64, nn int, direct bool) {

	var n, mmax, m, j, istep, i int
	var wtemp, wr, wpr, wpi, wi, theta, sign float64
	var tempr, tempi, temp float64

	if direct {
		sign = 1.0
	} else {
		sign = -1.0
	}

	n = nn << 1
	j = 1
	for i = 1; i < n; i += 2 {
		if j > i {
			/* swap data[j-1] and data[i-1]) */
			temp = data[j-1]
			data[j-1] = data[i-1]
			data[i-1] = temp

			/* swap data[j] and data[i]) */
			temp = data[j]
			data[j] = data[i]
			data[i] = temp
		}
		m = n >> 1
		for m >= 2 && j > m {
			j -= m
			m >>= 1
		}
		j += m
	}
	mmax = 2

	for n > mmax {

		istep = mmax << 1
		theta = sign * (6.28318530717959 / float64(mmax))
		wtemp = math.Sin(0.5 * theta)
		wpr = -2.0 * wtemp * wtemp
		wpi = math.Sin(theta)
		wr = 1.0
		wi = 0.0
		for m = 1; m < mmax; m += 2 {
			for i = m; i <= n; i += istep {
				j = i + mmax
				tempr = wr*data[j-1] - wi*data[j]
				tempi = wr*data[j] + wi*data[j-1]
				data[j-1] = data[i-1] - tempr
				data[j] = data[i] - tempi
				data[i-1] += tempr
				data[i] += tempi
			}
			wtemp = wr
			wr = wtemp*wpr - wi*wpi + wr
			wi = wi*wpr + wtemp*wpi + wi
		}
		mmax = istep
	}
}

/*
Compute DFT of a real discrete signal.
(Adapted fron Numerical Recipes Book)

Input array is the sequence of real values.

Output is stored in the same array using a strange scheme. The
first value is the Re{DFT[0]}, the second value is Re{DFT[N-1]}.
Example (all values rounded to first decimal):

  Real Input sequence N=16:
   0.5 1.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0

  Real DFT (rounded values):
   real[k] sum_n {inArray[n] * cos(alpha * k * n)}
   1.5 1.4 1.2 0.9 0.5 0.1 -0.2 -0.4 -0.5 -0.4 -0.2 0.1 0.5 0.9 1.2 1.4

  Imag DFT (rounded values):
   imag[k] sum_n {-inArray[n] * sin(alpha * k * n)}
   0.0 -0.4 -0.7 -0.9 -1.0 -0.9 -0.7 -0.4 0.0 0.4 0.7 0.9 1.0 0.9 0.7 0.4

  realft returns:
   1.5 -0.5 1.4 0.4 1.2 0.7 0.9 0.9 0.5 1.0 0.1 0.9 -0.2 0.7 -0.4 0.4
   Re   Re  Re  Im  Re  Im  Re  Im  Re  Im  Re  Im   Re  Im   Re  Im
   n=0  n=8 n=7 n=7 n=6 n=6 n=5 n=5 n=4 n=4 n=3 n=3  n=2 n=2  n=1 n=1
   The first 2 components are real values. The rest of the pairs are {Re, Im}

data is the input array of length n.
n the length of the discrete signal.
direct=true for direct FFT. Inverse otherwise.
*/
func RealFT(data []float64, n int, direct bool) {

	var i, i1, i2, i3, i4, np3 int
	var c1, c2, h1r, h1i, h2r, h2i float64
	var wr, wi, wpr, wpi, wtemp, theta float64

	c1 = 0.5
	theta = 3.141592653589793 / float64(n>>1)
	if direct {
		c2 = -0.5
		four1(data, n>>1, true)
	} else {
		c2 = 0.5
		theta = -theta
	}
	wtemp = math.Sin(0.5 * theta)
	wpr = -2.0 * wtemp * wtemp
	wpi = math.Sin(theta)
	wr = 1.0 + wpr
	wi = wpi
	np3 = n + 3
	for i = 2; i <= (n >> 2); i++ {
		i1 = i + i - 1
		i2 = 1 + i1
		i3 = np3 - i2
		i4 = 1 + i3
		h1r = c1 * (data[i1-1] + data[i3-1])
		h1i = c1 * (data[i2-1] - data[i4-1])
		h2r = -c2 * (data[i2-1] + data[i4-1])
		h2i = c2 * (data[i1-1] - data[i3-1])
		data[i1-1] = h1r + wr*h2r - wi*h2i
		data[i2-1] = h1i + wr*h2i + wi*h2r
		data[i3-1] = h1r - wr*h2r + wi*h2i
		data[i4-1] = -h1i + wr*h2i + wi*h2r
		wtemp = wr
		wr = wtemp*wpr - wi*wpi + wr
		wi = wi*wpr + wtemp*wpi + wi
	}
	h1r = data[0]
	if direct {
		data[0] = h1r + data[1]
		data[1] = h1r - data[1]
	} else {
		data[0] = c1 * (h1r + data[1])
		data[1] = c1 * (h1r - data[1])
		four1(data, n>>1, false)
	}
}

/*
Compute DFT energy vector.
The size of the energy array should be half of the input array.

     For the example in RealFT, the output would be:

     DFT Energy: 2.25 2.17 1.96 1.63 1.25 0.87 0.54 0.33
                 n=0  n=1  n=2  n=3  n=4  n=5  n=6  n=7

dft is the discrete Fourier transform. (See RealfFT for format.)
energy is the energy values for the DFT.
*/
func DFTEnergy(dft []float64) []float64 {

	size := len(dft) / 2
	energy := make([]float64, size, size)

	energy[0] = dft[0] * dft[0]
	for i := 1; i < size; i++ {
		energy[i] = dft[2*i]*dft[2*i] + dft[2*i+1]*dft[2*i+1]
	}
	return energy
}

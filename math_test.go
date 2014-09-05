package dsp

import "testing"

/*
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
*/

func TestRealFT(t *testing.T) {

	//   Real Input sequence N=16:
	//   0.5 1.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0
	data := make([]float64, 16, 16)
	data[0] = 0.5
	data[1] = 1.0

	//expected = []flaot64{1.5, 1.4, 1.2, 0.9, 0.5, 0.1, -0.2, -0.4, -0.5, -0.4, -0.2, 0.1, 0.5, 0.9, 1.2, 1.4}
	expected := []float64{1.5, -0.5, 1.4, 0.4, 1.2, 0.7, 0.9, 0.9, 0.5, 1.0, 0.1, 0.9, -0.2, 0.7, -0.4, 0.4}

	RealFT(data, 16, true)

	CompareSliceFloat(t, expected, data, "no match", 0.05)

}

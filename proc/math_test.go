package proc

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

	expected := []float64{1.5, -0.5, 1.4, 0.4, 1.2, 0.7, 0.9, 0.9, 0.5, 1.0, 0.1, 0.9, -0.2, 0.7, -0.4, 0.4}

	RealFT(data, 16, true)

	compareSliceFloat(t, expected, data, "no match", 0.05)
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
func TestDFTEnergy(t *testing.T) {

	//   Real Input sequence N=16:
	//   0.5 1.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0 0.0
	data := make([]float64, 16, 16)
	data[0] = 0.5
	data[1] = 1.0
	RealFT(data, 16, true)
	dft := data // in place
	energy := DFTEnergy(dft)

	expected := []float64{2.25, 2.17, 1.96, 1.63, 1.25, 0.87, 0.54, 0.33}

	compareSliceFloat(t, expected, energy, "no match", 0.05)

}

func TestGenerateFilterbank(t *testing.T) {

	indices, coeff := GenerateFilterbank(32, 6)

	t.Log(indices)
	t.Log(coeff)

	indices, coeff = GenerateFilterbank(1024, 10, 150, 1, 45)

	t.Log(indices)
	t.Log(coeff)

}

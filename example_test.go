package dsp_test

import (
	"fmt"

	"github.com/akualab/dsp"
	"github.com/akualab/dsp/wav"

	narray "github.com/akualab/narray/na64"
)

func ExampleApp_Spectrum() {

	app := dsp.NewApp("Example App")

	// Read a waveform.
	path := "data/wav1.json.gz"
	wavSource, err := wav.NewSourceProc(path, wav.Fs(8000))
	if err != nil {
		panic(err)
	}

	// Add the source processor responsible for reading and supplying waveform samples.
	wav := app.Add("wav", wavSource)

	// Use a windowing processor to segment the waveform into frames of 80 samples
	// and apply a Hamming window of size 205. The last arg is to instruct the processor
	// to center the frame in the middle of the window.
	window := app.Add("window", dsp.NewWindowProc(80, 205, dsp.Hamming, true))

	// Compute the FFT of the windowed frame. The FFT size is 2**8.
	spectrum := app.Add("spectrum", dsp.SpectralEnergy(8))

	// Connect the processors.
	// wav -> window -> spectrum
	app.Connect(window, wav)
	app.Connect(spectrum, window)

	// Get features using this object.
	out := spectrum

	// Get the next waveform. (This processor is designed to read a list of
	// files. The Next() method loads the next waveform in the list.)
	wavSource.Next()

	// Get the spectrum for frame #10.
	v, e := out.Get(10)
	if e != nil {
		panic(e)
	}

	// Print first element of the FFT for frame #10.
	fmt.Println(v.Data[0])
	// Output:
	// 5.123420120893221e-05
}

// Calculate Fibonacci values.
func ExampleApp_Fibonacci() {
	app := dsp.NewApp("Fibonacci")
	fibo := app.Add("Fibo", Fibo(11))

	// Self loop.
	app.Connect(fibo, fibo)

	f10, e := fibo.Get(10)
	if e != nil {
		panic(e)
	}
	fmt.Println(f10.Data[0])
	// Output:
	// 89
}

// Subtract a value "fm" from a Fibonacci series so the sum adds up to zero.
func ExampleOneValuer_ZMFibonacci() {
	app := dsp.NewApp("Fibonacci")
	fibo := app.Add("Fibo", Fibo(11))

	// Computes the mean of the series.
	fiboMean := app.Add("Fibo Mean", dsp.Mean())

	// Subtracts mean from fibo values.
	fiboZM := app.Add("ZM Fibo", dsp.Sub())

	// Make connections. Note that fiboMean is of tyep OneValuer and fibo is a Framer.
	// The reulting fiboZM is of type Framer. Values are computed only once and saved
	// in the processor cache.
	app.Connect(fibo, fibo)
	app.Connect(fiboMean, fibo)
	app.Connect(fiboZM, fibo, fiboMean)

	f10, e := fiboZM.Get(10)
	if e != nil {
		panic(e)
	}
	fmt.Println(f10.Data[0])
	// Output:
	// 67.9090909090909
}

// Fibo is a processor that returns the Fibonacci value for index.
// N is the length of the series.
func Fibo(N int) dsp.Processer {
	return dsp.NewProc(20, func(idx int, in ...dsp.Processer) (dsp.Value, error) {
		if idx == 0 || idx == 1 {
			na := narray.New(1)
			na.Set(1, 0)
			return na, nil
		}
		if idx < 0 || idx >= N {
			return nil, dsp.ErrOOB
		}
		vm1, err := dsp.Processers(in).Get(idx - 1)
		if err != nil {
			return nil, err
		}
		vm2, err := dsp.Processers(in).Get(idx - 2)
		if err != nil {
			return nil, err
		}

		res := narray.Add(nil, vm2, vm1)
		return res, nil
	})
}

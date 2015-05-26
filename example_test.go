package dsp_test

import (
	"fmt"

	"github.com/akualab/dsp"
	"github.com/akualab/dsp/wav"
)

func ExampleSpectrum() {

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

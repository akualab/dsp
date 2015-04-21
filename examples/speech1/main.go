package main

import (
	"log"
	"os"

	"github.com/akualab/dsp"
)

/* Configuration parameters. */
const (
	windowSize     = 205
	windowStep     = 80
	logFFTSize     = 8
	filterbankSize = 18
	cepstrumSize   = 8
)

func main() {

	app := dsp.NewApp("Test Chain", 1000)

	r, err := os.Open("data/audio-rec-8k.txt")
	if err != nil {
		panic(err)
	}

	c := &dsp.ReaderConfig{
		FrameSize: windowSize,
		StepSize:  windowStep,
		ValueType: dsp.Text,
	}

	print := true

	// Create a processor to compute mel-cepstrum. We chain various processors.
	cepstrum := app.Pipeline(
		dsp.Reader(r, c),
		dsp.Window(windowSize).Use(dsp.Hamming),
		dsp.SpectralEnergy(logFFTSize),
		dsp.Filterbank(dsp.MelFilterbankIndices, dsp.MelFilterbankCoefficients),
		dsp.Log(),
		dsp.DCT(filterbankSize, cepstrumSize),
		dsp.WriteValues(os.Stdout, print),
	)

	out := app.Run(
		cepstrum,
	)

	if app.Error() != nil {
		log.Fatalf("error: %s", app.Error())
	}

	// get a vector
	for v := range out {
		_ = v
	}

	err = r.Close()
	if err != nil {
		panic(err)
	}
}

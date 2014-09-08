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
	maxNormAlpha   = 0.99
	cepMeanWin     = 100
)

// An example of a full front-end implementation for speech recognition.
// We use the builder functions to make the app graph.
// The next step would be to generate the app graph from a json file.
// I leave it as an exercise for the reader :-)
func main() {

	app := dsp.NewApp("Speech Front-End", 1000)

	r, err := os.Open("data/audio-rec-8k.txt")
	if err != nil {
		panic(err)
	}

	c := &dsp.ReaderConfig{
		FrameSize: windowSize,
		StepSize:  windowStep,
		ValueType: dsp.Text,
	}

	// Use builder to create the application graph.
	b := app.NewBuilder()

	// Turns off writer output.
	print := false

	b.Add("waveform", dsp.Reader(r, c))
	b.Add("windowed", dsp.Window(windowSize).Use(dsp.Hamming))
	b.Add("spectrum", dsp.SpectralEnergy(logFFTSize))
	b.Add("filterbank", dsp.Filterbank(dsp.MelFilterbankIndices, dsp.MelFilterbankCoefficients))
	b.Add("log filterbank", dsp.Log())
	b.Add("cepstrum", dsp.DCT(filterbankSize, cepstrumSize))
	b.Tap("cepstrum")
	b.Add("writer", dsp.WriteValues(os.Stdout, print))
	b.Add("mean cepstrum", dsp.MovingAverage(cepstrumSize, cepMeanWin, nil))
	b.Tap("mean cepstrum")

	b.Connect("waveform", "windowed")
	b.Connect("windowed", "spectrum")
	b.Connect("spectrum", "filterbank")
	b.Connect("filterbank", "log filterbank")
	b.Connect("log filterbank", "cepstrum")
	b.Connect("cepstrum", "writer")
	b.Connect("cepstrum", "mean cepstrum")

	b.Add("cepstral energy", dsp.Sum())
	b.Tap("cepstral energy")
	b.Add("max cepstral energy", dsp.MaxNorm(maxNormAlpha))
	b.Tap("max cepstral energy")
	b.Connect("log filterbank", "cepstral energy")
	b.Connect("cepstral energy", "max cepstral energy")

	// Run the app.
	b.Run()

	// Get the output channels.
	// Must be done after calling Run()
	cepChan := b.TapChan("cepstrum")
	meanCepChan := b.TapChan("mean cepstrum")
	cepEgyChan := b.TapChan("cepstral energy")
	maxCepEgyChan := b.TapChan("max cepstral energy")

	if app.Error() != nil {
		log.Fatalf("error: %s", app.Error())
	}

	// get a vector
	i := 0
	for v := range cepChan {
		log.Printf("%5d: %v", i, v)
		vv := <-meanCepChan
		log.Printf("%5d: %v", i, vv)
		w := <-cepEgyChan
		log.Printf("%5d: %v", i, w)
		x := <-maxCepEgyChan
		log.Printf("%5d: %v", i, x)

		i++
	}

	err = r.Close()
	if err != nil {
		panic(err)
	}
}

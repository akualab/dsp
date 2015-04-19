package main

import (
	"bytes"
	"fmt"
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

var deltaCoeff = []float64{0.7, 0.2, 0.1}

// An example of a full front-end implementation for speech recognition.
// The audio data sampling rate is 8 KHz.
// We show how to use the builder functions to make the app graph.
// The next step would be to generate the app graph from a json file.
// I leave it as an exercise for the reader :-)
func main() {

	app := dsp.NewApp("Speech Recognizer Front-End", 1000)

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
	b.Add("writer", dsp.WriteValues(os.Stdout, print))
	b.Add("mean cepstrum", dsp.MovingAverage(cepstrumSize, cepMeanWin, nil))
	b.Add("zero mean cepstrum", dsp.Sub())

	b.Connect("waveform", "windowed")
	b.Connect("windowed", "spectrum")
	b.Connect("spectrum", "filterbank")
	b.Connect("filterbank", "log filterbank")
	b.Connect("log filterbank", "cepstrum")
	b.Connect("cepstrum", "writer")
	b.Connect("cepstrum", "mean cepstrum")

	// Subtract approx. cepstrum mean from cepstrum.
	b.ConnectOrdered("cepstrum", "zero mean cepstrum", 0)
	b.ConnectOrdered("mean cepstrum", "zero mean cepstrum", 1)

	// Energy features.
	b.Add("cepstral energy", dsp.Sum())
	b.Add("max cepstral energy", dsp.MaxNorm(maxNormAlpha))
	b.Connect("log filterbank", "cepstral energy")
	b.Connect("cepstral energy", "max cepstral energy")

	// Subtract approx. max energy from energy.
	b.Add("normalized cepstral energy", dsp.Sub())
	b.ConnectOrdered("cepstral energy", "normalized cepstral energy", 0)
	b.ConnectOrdered("max cepstral energy", "normalized cepstral energy", 1)

	// Delta cepstrum features.
	b.Add("delta cepstrum", dsp.NewDiffProc(cepstrumSize, deltaCoeff))
	b.Add("delta delta cepstrum", dsp.NewDiffProc(cepstrumSize, deltaCoeff))
	b.Connect("zero mean cepstrum", "delta cepstrum")
	b.Connect("delta cepstrum", "delta delta cepstrum")

	// Delta energy features.
	b.Add("delta energy", dsp.NewDiffProc(1, deltaCoeff))
	b.Add("delta delta energy", dsp.NewDiffProc(1, deltaCoeff))
	b.Connect("normalized cepstral energy", "delta energy")
	b.Connect("delta energy", "delta delta energy")

	// Put three cepstrum features and three energy features in a single vector.
	b.Add("combined", dsp.Join())
	b.Tap("combined")
	b.ConnectOrdered("normalized cepstral energy", "combined", 0)
	b.ConnectOrdered("delta energy", "combined", 1)
	b.ConnectOrdered("delta delta energy", "combined", 2)
	b.ConnectOrdered("zero mean cepstrum", "combined", 3)
	b.ConnectOrdered("delta cepstrum", "combined", 4)
	b.ConnectOrdered("delta delta cepstrum", "combined", 5)

	// Run the app. Here is where the channels are created and assigned to processors.
	b.Run()

	// Get the output channels.
	// Must be done after calling Run()
	combChan := b.TapChan("combined")

	// Get vectors.
	i := 0
	for v := range combChan {
		fmt.Printf("f=%5d, len=%d, v=%v\n", i, len(v), vecSprint(v))
		i++
	}

	if app.Error() != nil {
		log.Fatalf("error: %s", app.Error())
	}

	err = r.Close()
	if err != nil {
		panic(err)
	}
}

func vecSprint(vec []float64) string {

	var buffer bytes.Buffer

	for _, v := range vec {
		s := fmt.Sprintf("%7.1f ", v)
		_, _ = buffer.WriteString(s)
	}
	return buffer.String()
}

package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/akualab/dsp"
	"github.com/akualab/dsp/wav"
)

/* Configuration parameters. */
const (
	bufSize        = 100
	windowSize     = 205
	windowStep     = 80
	logFFTSize     = 8
	filterbankSize = 18
	cepstrumSize   = 8
	maxNormAlpha   = 0.99
	cepMeanWin     = 100
	fs             = 8000
	path           = "../../data"
)

var deltaCoeff = []float64{0.7, 0.2, 0.1}

// An example of a full front-end implementation for speech recognition.
// The audio data sampling rate is 8 KHz.
// We show how to use the builder functions to make the app graph.
// The next step would be to generate the app graph from a json file.
// I leave it as an exercise for the reader :-)
func main() {

	app := dsp.NewApp("Speech Recognizer Front-End")

	wavSource, err := wav.NewSourceProc(path, fs, windowSize, windowStep, false)
	if err != nil {
		log.Fatal(err)
	}

	app.Add("waveform", wavSource)
	app.Add("windowed", dsp.Window(windowSize).Use(dsp.Hamming))
	app.Add("spectrum", dsp.SpectralEnergy(logFFTSize))
	app.Add("filterbank", dsp.Filterbank(dsp.MelFilterbankIndices, dsp.MelFilterbankCoefficients))
	app.Add("log filterbank", dsp.Log())
	app.Add("cepstrum", dsp.DCT(filterbankSize, cepstrumSize))
	app.Add("mean cepstrum", dsp.NewMAProc(cepstrumSize, cepMeanWin, bufSize))
	app.Add("zero mean cepstrum", dsp.Sub())

	app.Connect("windowed", "waveform")
	app.Connect("spectrum", "windowed")
	app.Connect("filterbank", "spectrum")
	app.Connect("log filterbank", "filterbank")
	app.Connect("cepstrum", "log filterbank")
	app.Connect("mean cepstrum", "cepstrum")

	// mean cep uses two inputs
	app.Connect("zero mean cepstrum", "cepstrum", "mean cepstrum")

	// Energy features.
	app.Add("cepstral energy", dsp.Sum())
	app.Add("max cepstral energy", dsp.MaxNorm(bufSize, maxNormAlpha))
	app.Connect("cepstral energy", "log filterbank")
	app.Connect("max cepstral energy", "cepstral energy")

	// Subtract approx. max energy from energy.
	app.Add("normalized cepstral energy", dsp.Sub())
	app.Connect("normalized cepstral energy", "cepstral energy", "max cepstral energy")

	// Delta cepstrum features.
	app.Add("delta cepstrum", dsp.NewDiffProc(cepstrumSize, bufSize, deltaCoeff))
	app.Add("delta delta cepstrum", dsp.NewDiffProc(cepstrumSize, bufSize, deltaCoeff))
	app.Connect("delta cepstrum", "zero mean cepstrum")
	app.Connect("delta delta cepstrum", "delta cepstrum")

	// Delta energy features.
	app.Add("delta energy", dsp.NewDiffProc(1, bufSize, deltaCoeff))
	app.Add("delta delta energy", dsp.NewDiffProc(1, bufSize, deltaCoeff))
	app.Connect("delta energy", "normalized cepstral energy")
	app.Connect("delta delta energy", "delta energy")

	// Put three cepstrum features and three energy features in a single vector.
	app.Add("combined", dsp.Join())
	out := app.NewTap("combined")

	app.Connect("combined",
		"normalized cepstral energy",
		"delta energy",
		"delta delta energy",
		"zero mean cepstrum",
		"delta cepstrum",
		"delta delta cepstrum")

	for {
		// load next wav
		err := wavSource.Next()
		if err == wav.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		numFrames := wavSource.NumFrames()
		id := wavSource.ID()
		if numFrames == 0 {
			log.Printf("waveform %s is too short, skipping", id)
			continue
		}
		app.Reset()
		log.Printf("processing waveform [%s] with %d frames, mean: %6.2f, sd: %6.2f", id, numFrames, wavSource.Mean(), wavSource.SD())
		var i uint32
		for ; ; i++ {
			v, e := out.Get(i)
			if e == dsp.ErrOOB {
				log.Printf("done processing %d frames for waveform [%s]", i, id)
				break
			}
			if e != nil {
				log.Fatal(e)
			}
			log.Printf("feature: cepstrum, frame: %d, data: %v", i, v.Data)
		}
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

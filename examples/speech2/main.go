package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/akualab/dsp"
	"github.com/akualab/dsp/proc"
	"github.com/akualab/dsp/proc/speech"
	"github.com/akualab/dsp/proc/wav"
	narray "github.com/akualab/narray/na64"
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

	c := speech.Config{
		FS:         8000,
		BufSize:    100,
		WinSize:    205,
		WinStep:    80,
		WinType:    proc.Hamming,
		LogFFTSize: 8,
		FBSize:     18,
		FBMinFreq:  10,
		FBMaxFreq:  3500,
		CepSize:    8,
		DeltaCoeff: deltaCoeff,
	}

	wavSource, err := wav.NewSourceProc(path, wav.Fs(fs))
	if err != nil {
		log.Fatal(err)
	}
	app, err := speech.New("Speech Recognizer Front-End", wavSource, c)
	if err != nil {
		log.Fatalf("can't init speech app, error: %s", err)
	}
	out := app.NodeByName("combined")

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
		for i := 0; ; i++ {
			v, e := out.Get(i)
			if e == dsp.ErrOOB {
				log.Printf("done processing %d frames for waveform [%s]", i, id)
				break
			}
			if e != nil {
				log.Fatal(e)
			}
			log.Printf("feature: %s, frame: %d, data: %v", out.Name(), i, v.(*narray.NArray).Data)
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

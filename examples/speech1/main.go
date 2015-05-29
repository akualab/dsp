package main

import (
	"log"

	"github.com/akualab/dsp"
	"github.com/akualab/dsp/proc"
	"github.com/akualab/dsp/proc/speech"
	"github.com/akualab/dsp/proc/wav"
	narray "github.com/akualab/narray/na64"
)

/* Configuration parameters. */
const (
	windowSize     = 205 // frameSize
	windowStep     = 80  // stepSize
	logFFTSize     = 8
	filterbankSize = 18
	cepstrumSize   = 8
	fs             = 8000
	path           = "../../data"
)

func main() {

	app := dsp.NewApp("Test Chain")

	wavSource, err := wav.NewSourceProc(path, wav.Fs(fs))
	if err != nil {
		log.Fatal(err)
	}

	out := app.Chain(
		app.Add("cepstrum", proc.DCT(filterbankSize, cepstrumSize)),
		app.Add("log_filterbank", proc.Log()),
		app.Add("filterbank", proc.Filterbank(speech.MelFilterbankIndices, speech.MelFilterbankCoefficients)),
		app.Add("spectral_energy", proc.SpectralEnergy(logFFTSize)),
		app.Add("window", proc.NewWindowProc(windowStep, windowSize, proc.Hamming, false)),
		app.Add("wav", wavSource),
	)

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
		log.Printf("processing waveform [%s] with %d frames", id, numFrames)
		for i := 0; ; i++ {
			v, e := out.Get(i)
			if e == dsp.ErrOOB {
				log.Printf("done processing %d frames for waveform [%s]", i, id)
				break
			}
			if e != nil {
				log.Fatal(e)
			}
			log.Printf("feature: cepstrum, frame: %d, data: %v", i, v.(*narray.NArray).Data)
		}
	}
}

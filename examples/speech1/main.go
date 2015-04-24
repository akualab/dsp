package main

import (
	"log"

	"github.com/akualab/dsp"
	"github.com/akualab/dsp/wav"
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

	wavSource, err := wav.NewSourceProc(path, fs, windowSize, windowStep, false)
	if err != nil {
		log.Fatal(err)
	}

	app.Add("wav", wavSource)
	app.Add("window", dsp.Window(windowSize).Use(dsp.Hamming))
	app.Add("spectral_energy", dsp.SpectralEnergy(logFFTSize))
	app.Add("filterbank", dsp.Filterbank(dsp.MelFilterbankIndices, dsp.MelFilterbankCoefficients))
	app.Add("log_filterbank", dsp.Log())
	app.Add("cepstrum", dsp.DCT(filterbankSize, cepstrumSize))

	app.Connect("window", "wav")
	app.Connect("spectral_energy", "window")
	app.Connect("filterbank", "spectral_energy")
	app.Connect("log_filterbank", "filterbank")
	app.Connect("cepstrum", "log_filterbank")

	out := app.NewTap("cepstrum")

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

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
	fftSize        = 256
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
	out := app.Run(
		dsp.Reader(r, c),                        // Read audio data from file.
		dsp.Window(windowSize).Use(dsp.Hamming), // Applies Hamming window to frame.
		dsp.WriteValues(os.Stdout, print),       // Writes to stdout.
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

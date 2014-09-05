package main

import (
	"log"
	"os"

	"github.com/akualab/dsp"
)

func main() {

	app := dsp.NewApp("Test Chain", 1000)

	out := app.Run(
		dsp.Source(64, 2).Use(dsp.NewSquare(1, 0, 4, 4)),
		dsp.Window(64).Use(dsp.Hamming),
		dsp.WriteValues(os.Stdout, true),
	)

	if app.Error() != nil {
		log.Fatalf("error: %s", app.Error())
	}

	// get a vector
	_ = <-out

}

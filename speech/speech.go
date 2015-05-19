// Package speech provides functionality to parametrize digital waveforms. It computes cepstral features
// using a sequence of short-term discrete Fourier transforms. Log filterbanks are computed from teh DFT
// and finally the cepstrum is computed using the discrete cosine transform.
//
// Originally developed for speech recognition, this representation can be used as the basis for other
// applications.
//
// This package was implemented using package github.com/akualab/dsp which should make it wasy to modify
// and adapt to any application.
package speech

import (
	"github.com/akualab/dsp"
	"github.com/akualab/dsp/wav"
)

// Config parameters for speech feature extractor.
type Config struct {
	// Sampling rate.
	FS float64
	// Processsor buffer size.
	BufSize int
	// Frame size in samples.
	WinSize int
	// Frame advance step in samples.
	WinStep int
	// Window Type (0: Rect, 1: Hann, 2: Hamm, 3: Blackman)
	WinType int
	// Log of the FFT size in samples.
	LogFFTSize int
	// Number fo filterbank elements.
	FBSize int
	// Min filterbank frequency.
	FBMinFreq float64
	// Max filterbank frequency.
	FBMaxFreq float64
	// Number of cepstral elements.
	CepSize int
	// Coefficients for computing deltas.
	DeltaCoeff []float64
	// Name of the feature(s).
	Features []string
	// Dim is the dimension of the feature vector.
	Dim int
}

// DefaultFeatures has a list of the default feature names.
var DefaultFeatures = []string{
	"normalized cepstral energy",
	"delta energy",
	"delta delta energy",
	"zm cepstrum",
	"delta cepstrum",
	"delta delta cepstrum",
}

// New creates a new speech dsp app.
func New(name string, source *wav.SourceProc, c Config) (*dsp.App, error) {

	app := dsp.NewApp(name)
	app.Add("wav", source)
	app.Add("windowed", dsp.NewWindowProc(c.WinStep, c.WinSize, c.WinType, true))
	app.Add("spectrum", dsp.SpectralEnergy(c.LogFFTSize))
	indices, coeff := dsp.GenerateFilterbank(1<<uint(c.LogFFTSize), c.FBSize, c.FS, c.FBMinFreq, c.FBMaxFreq)
	app.Add("filterbank", dsp.Filterbank(indices, coeff))
	app.Add("log filterbank", dsp.Log())
	app.Add("cepstrum", dsp.DCT(c.FBSize, c.CepSize))
	app.Add("mean cepstrum", dsp.Mean())
	app.Add("zm cepstrum", dsp.Sub(true))

	app.Connect("windowed", "wav")
	app.Connect("spectrum", "windowed")
	app.Connect("filterbank", "spectrum")
	app.Connect("log filterbank", "filterbank")
	app.Connect("cepstrum", "log filterbank")
	app.Connect("mean cepstrum", "cepstrum")
	app.Connect("zm cepstrum", "cepstrum", "mean cepstrum")

	// Energy features.
	app.Add("cepstral energy", dsp.Sum())
	app.Add("max cepstral energy", dsp.MaxWin())
	app.Connect("cepstral energy", "log filterbank")
	app.Connect("max cepstral energy", "cepstral energy")

	// Subtract max energy from energy.
	app.Add("normalized cepstral energy", dsp.Sub(true))
	app.Connect("normalized cepstral energy", "cepstral energy", "max cepstral energy")

	// Delta cepstrum features.
	app.Add("delta cepstrum", dsp.NewDiffProc(c.CepSize, c.BufSize, c.DeltaCoeff))
	app.Add("delta delta cepstrum", dsp.NewDiffProc(c.CepSize, c.BufSize, c.DeltaCoeff))
	app.Connect("delta cepstrum", "zm cepstrum")
	app.Connect("delta delta cepstrum", "delta cepstrum")

	// Delta energy features.
	app.Add("delta energy", dsp.NewDiffProc(1, c.BufSize, c.DeltaCoeff))
	app.Add("delta delta energy", dsp.NewDiffProc(1, c.BufSize, c.DeltaCoeff))
	app.Connect("delta energy", "normalized cepstral energy")
	app.Connect("delta delta energy", "delta energy")

	// Put three energy features and cepstrum features in a single vector.
	app.Add("combined", dsp.Join())
	app.Connect("combined", c.Features...)
	return app, nil
}

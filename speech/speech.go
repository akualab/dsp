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

	if len(c.Features) == 0 {
		c.Features = DefaultFeatures
	}
	app := dsp.NewApp(name)
	indices, coeff := dsp.GenerateFilterbank(1<<uint(c.LogFFTSize), c.FBSize, c.FS, c.FBMinFreq, c.FBMaxFreq)

	cep := app.Chain(
		app.Add("cepstrum", dsp.DCT(c.FBSize, c.CepSize)),
		app.Add("log filterbank", dsp.Log()),
		app.Add("filterbank", dsp.Filterbank(indices, coeff)),
		app.Add("spectrum", dsp.SpectralEnergy(c.LogFFTSize)),
		app.Add("windowed", dsp.NewWindowProc(c.WinStep, c.WinSize, c.WinType, true)),
		app.Add("wav", source),
	)

	meanCep := app.Connect(
		app.Add("mean cepstrum", dsp.Mean()),
		cep,
	)

	zmCep := app.Connect(
		app.Add("zm cepstrum", dsp.Sub(true)),
		cep,
		meanCep,
	)

	// Energy features.
	egy := app.Connect(
		app.Add("cepstral energy", dsp.Sum()),
		app.NodeByName("log filterbank"),
	)

	maxEgy := app.Connect(
		app.Add("max cepstral energy", dsp.MaxWin()),
		egy,
	)

	// Subtract max energy from energy.
	normEgy := app.Connect(
		app.Add("normalized cepstral energy", dsp.Sub(true)),
		egy,
		maxEgy,
	)

	// Delta cepstrum features.
	dCep := app.Connect(
		app.Add("delta cepstrum", dsp.NewDiffProc(c.CepSize, c.BufSize, c.DeltaCoeff)),
		zmCep,
	)
	app.Connect(
		app.Add("delta delta cepstrum", dsp.NewDiffProc(c.CepSize, c.BufSize, c.DeltaCoeff)),
		dCep,
	)

	// Delta energy features.
	dEgy := app.Connect(
		app.Add("delta energy", dsp.NewDiffProc(1, c.BufSize, c.DeltaCoeff)),
		normEgy,
	)
	app.Connect(
		app.Add("delta delta energy", dsp.NewDiffProc(1, c.BufSize, c.DeltaCoeff)),
		dEgy,
	)

	// Put three energy features and cepstrum features in a single vector.
	nodes, err := app.NodesByName(c.Features...)
	if err != nil {
		return nil, err
	}
	app.Connect(
		app.Add("combined", dsp.Join()),
		nodes...,
	)
	return app, nil
}

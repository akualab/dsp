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
		app.Add("zm cepstrum", dsp.Sub()),
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
		app.Add("normalized cepstral energy", dsp.Sub()),
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

var (
	// MelFilterbankIndices are the indices of the filters in the filterbank.
	MelFilterbankIndices = []int{10, 11, 14, 17, 20, 23, 27, 30, 33, 36, 40, 45, 50, 56, 62, 69, 76, 84}
	// MelFilterbankCoefficients is a hardcoded filterbank for the speech example.
	MelFilterbankCoefficients = [][]float64{
		[]float64{1.0, 1.0, 1.0, 1.0, 0.66, 0.33},
		[]float64{0.33, 0.66, 1.0, 1.0, 1.0, 1.0, 0.66, 0.33},
		[]float64{0.33, 0.66, 1.0, 1.0, 1.0, 1.0, 0.66, 0.33},
		[]float64{0.33, 0.66, 1.0, 1.0, 1.0, 1.0, 0.75, 0.5, 0.25},
		[]float64{0.33, 0.66, 1.0, 1.0, 1.0, 1.0, 1.0, 0.66, 0.33},
		[]float64{0.25, 0.5, 0.75, 1.0, 1.0, 1.0, 1.0, 0.66, 0.33},
		[]float64{0.33, 0.66, 1.0, 1.0, 1.0, 1.0, 0.66, 0.33},
		[]float64{0.33, 0.66, 1.0, 1.0, 1.0, 1.0, 0.75, 0.5, 0.25},
		[]float64{0.33, 0.66, 1.0, 1.0, 1.0, 1.0, 1.0, 0.8, 0.6, 0.4, 0.2},
		[]float64{0.25, 0.5, 0.75, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 0.8, 0.6, 0.4, 0.2},
		[]float64{0.2, 0.4, 0.6, 0.8, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 0.83, 0.66, 0.5, 0.33, 0.16},
		[]float64{0.2, 0.4, 0.6, 0.8, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 0.83, 0.66, 0.5, 0.33, 0.16},
		[]float64{0.16, 0.33, 0.5, 0.66, 0.83, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 0.85, 0.71, 0.57, 0.42, 0.28, 0.14},
		[]float64{0.16, 0.33, 0.5, 0.66, 0.83, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 0.85, 0.71, 0.57, 0.42, 0.28, 0.14},
		[]float64{0.14, 0.28, 0.42, 0.57, 0.71, 0.85, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 0.875, 0.75, 0.625, 0.5, 0.375, 0.25, 0.125},
		[]float64{0.142, 0.285, 0.428, 0.571, 0.714, 0.857, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 0.88, 0.77, 0.66, 0.55, 0.44, 0.33, 0.22, 0.11},
		[]float64{0.125, 0.25, 0.375, 0.5, 0.625, 0.75, 0.875, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 0.88, 0.77, 0.66, 0.55, 0.44, 0.33, 0.22, 0.11},
		[]float64{0.11, 0.22, 0.33, 0.44, 0.55, 0.66, 0.77, 0.88, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0},
	}
)

// Copyright (c) 2015 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wav

import (
	"errors"
	"fmt"

	"github.com/akualab/dsp"
	"github.com/akualab/ju"
	narray "github.com/akualab/narray/na64"
)

// Done is returned as the error value when there are no more waveforms available in the stream.
var Done = errors.New("no more json objects")

// A Waveform format for reading json files.
//go:generate optioner -type Waveform
type Waveform struct {
	// ID is a waveform identifier.
	ID string `json:"id" opt:"-"`
	// Samples are the digital samples read as a float64.
	Samples []float64 `samples:"id" opt:"-"`
	// FS is the sampling frequency in Hertz.
	FS float64 `json:"fs,omitempty"`

	frameSize int
	stepSize  int
	idx       int     `opt:"-"`
	sumx      float64 `opt:"-"`
	sumxsq    float64 `opt:"-"`
}

// New returns a waveform object.
// The waveform is partitioned into frames that may overlap with each other.
// The length of each frame is set with option FrameSize().
// The distance between succesive frames is set with option StepSize.
// To get a single frame form the entire waveform use FrameSize(0) which is the default.
// If FrameSize equals the StepSize, the waveform is partitioned using disjoint segments.
// To specifiy a sampling rate, use option FS(). (In the future the package will convert the sampling rate.)
func New(id string, samples []float64, options ...optWaveform) Waveform {

	w := Waveform{
		ID:      id,
		Samples: samples,
	}

	// Set options.
	w.Option(options...)

	if w.frameSize < 1 {
		w.frameSize = len(samples)
		w.stepSize = len(samples)
	}
	w.stats()
	return w
}

// Iter is an iterator to access waveforms sequentially.
type Iter struct {
	js        *ju.JSONStreamer
	frameSize int
	stepSize  int
	fs        float64
}

// NewIterator creates an iterator to access all waveforms in path.
// Also see New() for more details.
// To specify path see ju.JSONStreamer.
// It is the caller's responsibility to call Close to release the underlying readers.
func NewIterator(path string, fs float64, frameSize, stepSize int) (Iter, error) {
	js, err := ju.NewJSONStreamer(path)
	if err != nil {
		return Iter{}, err
	}
	iter := Iter{
		js:        js,
		frameSize: frameSize,
		stepSize:  stepSize,
		fs:        fs,
	}
	return iter, nil
}

// Next returns the next available waveform.
// When there are no more waveforms, Done is returned as the error.
func (iter Iter) Next() (Waveform, error) {
	var w Waveform
	e := iter.js.Next(&w)
	if e == ju.Done {
		return w, Done
	}
	if w.FS > 0 && iter.fs > 0 && (w.FS != iter.fs) {
		fmt.Errorf("sampling rates don't match - wav fs is [%f], expected [%f] - TODO: implement sampling rate conversion", w.FS, iter.fs)
	}
	return New(w.ID, w.Samples, FrameSize(iter.frameSize), StepSize(iter.stepSize)), nil
}

// Close underlying readers.
func (iter Iter) Close() error {
	return iter.js.Close()
}

// NumFrames returns the maximum number of frames in the waveform.
func (w *Waveform) NumFrames() int {
	if w.stepSize < w.frameSize {
		return (len(w.Samples) - (w.frameSize - w.stepSize)) / w.stepSize
	}
	return len(w.Samples) / w.stepSize
}

// Frame returns a frame of samples for the given index. NOTE: the slice may be shared with
// other processors or may be cached. For these reason, the caller should not modify the slice in-place.
func (w *Waveform) Frame(idx int) (dsp.Value, error) {
	n := len(w.Samples)
	start := idx * w.stepSize
	end := start + w.frameSize
	if start < 0 || start >= n {
		return nil, dsp.ErrOOB
	}
	if end < 1 || end >= n {
		return nil, dsp.ErrOOB
	}
	res := narray.NewArray(w.Samples[start:end], w.frameSize)
	return res, nil
}

// Reset implements the dsp.Resetter interface.
// Sets frame index back to zero.
func (w *Waveform) Reset() {
	w.idx = 0
}

// NextFrame returns the next frame of samples sequentially.
// Returns error value Done when no additional frames are available.
// See also Frame().
func (w *Waveform) NextFrame(idx int) (dsp.Value, error) {
	frame, e := w.Frame(w.idx)
	if e == Done {
		return nil, Done
	}
	w.idx++
	return frame, e
}

// SourceProc is a source processor that provides access to waveform data.
type SourceProc struct {
	iter Iter
	wav  Waveform
	zm   bool
	*dsp.Proc
}

// NewSourceProc create a new source of waveforms.
// See also New() for more details.
// If zeroMean is true, the mean of the waveform samples is subtracetd from every sample.
func NewSourceProc(path string, fs float64, frameSize, stepSize int, zeroMean bool) (*SourceProc, error) {

	iter, err := NewIterator(path, fs, frameSize, stepSize)
	if err != nil {
		return nil, err
	}
	return &SourceProc{
		iter: iter,
		Proc: dsp.NewProc(100, nil),
		zm:   zeroMean,
	}, nil
}

// Next loads the next available waveform into the source. Returns Done when all waveforms have been processed.
func (src *SourceProc) Next() error {
	var err error
	src.wav, err = src.iter.Next()
	if err == Done {
		e := src.iter.Close()
		if e != nil {
			return e
		}
		return Done
	}
	if err != nil {
		return err
	}
	if src.zm {
		for i := range src.wav.Samples {
			src.wav.Samples[i] -= src.wav.Mean()
		}
	}
	return nil
}

// Get implements the dsp.Processer interface.
func (src *SourceProc) Get(idx uint32) (dsp.Value, error) {
	frame, err := src.wav.Frame(int(idx))
	if err != nil {
		return nil, err
	}
	return frame, nil
}

// ID returns the id of the current waveform.
func (src *SourceProc) ID() string {
	return src.wav.ID
}

// NumFrames returns the number of frames in the current waveform.
func (src *SourceProc) NumFrames() int {
	return src.wav.NumFrames()
}

// Mean returns the mean of the waveform samples.
func (src *SourceProc) Mean() float64 {
	return src.wav.Mean()
}

// SD returns the standard deviation of the waveform samples.
func (src *SourceProc) SD() float64 {
	return src.wav.SD()
}

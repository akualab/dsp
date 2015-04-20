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
)

// Done is returned as the error value when there are no more waveforms available in the stream.
var Done = errors.New("no more json objects")

// ErrOutOfBounds is returned when the frame index is out of bounds.
var ErrOutOfBounds = errors.New("frame index out of bounds")

// A Waveform format for reading json files.
type Waveform struct {
	// ID is a waveform identifier.
	ID string `json:"id"`
	// Samples are the digital samples read as a float64.
	Samples []float64 `samples:"id"`
	// FS is the sampling frequency in Hertz.
	FS float64 `json:"fs,omitempty"`

	frameSize int
	stepSize  int
	idx       int
}

// New returns a waveform object.
// The waveform is partitioned into frames that may overlap with each other.
// The length of each frame is the frameSize.
// The distance between succesive frames is the stepSize.
// If the frameSize equals the stepSize, the waveform is partitioned using disjoint segments.
func New(id string, samples []float64, fs float64, frameSize, stepSize int) (Waveform, error) {

	if stepSize < 1 {
		return Waveform{}, fmt.Errorf("step size is %d, must be greater than zero", stepSize)
	}
	if frameSize < 1 {
		return Waveform{}, fmt.Errorf("frame size is %d, must be greater than zero", frameSize)
	}
	w := Waveform{
		ID:        id,
		Samples:   samples,
		FS:        fs,
		frameSize: frameSize,
		stepSize:  stepSize,
	}
	return w, nil
}

// Iter is an iterator to access waveforms sequentially.
type Iter struct {
	js        *ju.JSONStreamer
	frameSize int
	stepSize  int
	fs        float64
}

// NewIterator creates an iterator to access all waveforms in path.
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
	if w.FS != iter.fs {
		fmt.Errorf("sampling rates don't match - wav fs is [%f], expected [%f] - TODO: implement sampling rate conversion", w.FS, iter.fs)
	}
	w.frameSize = iter.frameSize
	w.stepSize = iter.stepSize
	return w, e
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
		return nil, ErrOutOfBounds
	}
	if end < 1 || end >= n {
		return nil, ErrOutOfBounds
	}
	return dsp.Value(w.Samples[start:end]), nil
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

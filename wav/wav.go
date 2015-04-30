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
type Waveform struct {
	// ID is a waveform identifier.
	ID string `json:"id"`
	// Samples are the digital samples read as a float64.
	Samples []float64 `samples:"id"`
	// FS is the sampling frequency in Hertz.
	FS float64 `json:"fs,omitempty"`

	sumx   float64
	sumxsq float64
}

// New returns a waveform object.
// To specifiy a sampling rate, use option fs. Use fs=0 to ignore checks. (In the future the package will convert the sampling rate.)
func New(id string, samples []float64, fs float64) Waveform {

	w := Waveform{
		ID:      id,
		Samples: samples,
		FS:      fs,
	}

	w.stats()
	return w
}

// Iter is an iterator to access waveforms sequentially.
type Iter struct {
	js                           *ju.JSONStreamer
	frameSize, stepSize, winType int
	winData                      []float64
	fs                           float64
	wav                          Waveform
}

// NewIterator creates an iterator to access all waveforms in path.
// The waveform is partitioned into frames that may overlap with each other.
// The length of each frame is frameSize.
// The distance between succesive frames is stepSize.
// To get a single frame form the entire waveform use frameSize=0.
// If frameSize equals the stepSize, the waveform is partitioned using disjoint segments.
// To specify path see ju.JSONStreamer.
// It is the caller's responsibility to call Close to release the underlying readers.
func NewIterator(path string, fs float64, frameSize, stepSize int) (*Iter, error) {
	js, err := ju.NewJSONStreamer(path)
	if err != nil {
		return nil, err
	}
	iter := &Iter{
		js:        js,
		frameSize: frameSize,
		stepSize:  stepSize,
		fs:        fs,
	}

	return iter, nil
}

// Next returns the next available waveform.
// When there are no more waveforms, Done is returned as the error.
func (iter *Iter) Next() (Waveform, error) {
	return iter.NextSegment(0, -1)
}

// NextSegment returns a segment of the next available waveform.
// When there are no more waveforms, Done is returned as the error.
// Param start is the index of the start sample. (Must be less than len(wav) and end.)
// Param end is the max value of the index of the last sample to be included. (Must be less than len(wav).)
// To set end to the size of the waveform use end=-1.
func (iter *Iter) NextSegment(start, end int) (Waveform, error) {
	var w Waveform
	e := iter.js.Next(&w)
	if e == ju.Done {
		return w, Done
	}
	if w.FS > 0 && iter.fs > 0 && (w.FS != iter.fs) {
		fmt.Errorf("sampling rates don't match - wav fs is [%f], expected [%f] - TODO: implement sampling rate conversion", w.FS, iter.fs)
	}
	if iter.frameSize < 1 {
		iter.frameSize = len(w.Samples)
		iter.stepSize = iter.frameSize
	}
	iter.wav, e = getWav(w, start, end, iter.fs)
	return iter.wav, e
}

func getWav(w Waveform, start, end int, fs float64) (Waveform, error) {
	if start >= len(w.Samples) {
		fmt.Errorf("start must be less than length of wav, got start=%d, len(wav)=%d", start, len(w.Samples))
	}
	if end == -1 {
		end = len(w.Samples)
	}
	if start >= end {
		fmt.Errorf("start must be less than end, got start=%d, end=%d", start, end)
	}
	return New(w.ID, w.Samples[start:end], fs), nil
}

// Close underlying readers.
func (iter *Iter) Close() error {
	return iter.js.Close()
}

// NumFrames returns the maximum number of frames in the waveform.
func (iter *Iter) NumFrames() int {
	if iter.frameSize < 1 {
		return 1
	}
	if iter.stepSize < iter.frameSize {
		return (len(iter.wav.Samples) - (iter.frameSize - iter.stepSize)) / iter.stepSize
	}
	return len(iter.wav.Samples) / iter.stepSize
}

// Frame returns a frame of samples for the given index. NOTE: the slice may be shared with
// other processors or may be cached. For these reason, the caller should not modify the slice in-place.
func (iter *Iter) Frame(idx int) (dsp.Value, error) {
	n := len(iter.wav.Samples)
	start := idx * iter.stepSize
	end := start + iter.frameSize
	if start < 0 || start >= n {
		return nil, dsp.ErrOOB
	}
	if end < 1 || end > n {
		return nil, dsp.ErrOOB
	}
	res := narray.NewArray(iter.wav.Samples[start:end], iter.frameSize)
	return res, nil
}

// Reset implements the dsp.Resetter interface.
// Sets frame index back to zero.
func (w *Waveform) Reset() {
}

// SourceProc is a source processor that provides access to waveform data.
//go:generate optioner -type SourceProc
type SourceProc struct {
	*dsp.Proc  `opt:"-"`
	path       string
	iter       *Iter    `opt:"-"`
	wav        Waveform `opt:"-"`
	zm         bool
	winType    int
	winData    []float64 `opt:"-"`
	frameSize  int
	stepSize   int
	bufSize    int
	fs         float64
	start, end int
}

// NewSourceProc create a new source of waveforms.
// See also New() for more details.
// If zeroMean is true, the mean of the waveform samples is subtracetd from every sample.
// Note that calling Mean() will still return the original mean value. Think of Mean() as the original mean value.
func NewSourceProc(path string, options ...optSourceProc) (*SourceProc, error) {
	s := &SourceProc{path: path}

	// Set options.
	s.Option(options...)

	iter, err := NewIterator(path, s.fs, s.frameSize, s.stepSize)
	if err != nil {
		return nil, err
	}
	s.iter = iter
	s.Proc = dsp.NewProc(s.bufSize, nil)

	if s.winType > 0 {
		s.winData, err = dsp.WindowSlice(s.winType, s.frameSize)
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

// Rewind makes the the current waveform available for processing with different parameters.
// This is useful when the a waveform source needs to be segmented in multiple ways.
func (src *SourceProc) Rewind(start, end, frameSize, stepSize int, winType int) error {

	var err error
	if src.wav.Samples == nil {
		return fmt.Errorf("source proc has no waveform loaded, cannot rewind, call Next() before attempting to rewind")
	}

	src.iter.winType = winType
	src.iter.winData, err = dsp.WindowSlice(winType, frameSize)
	if err != nil {
		return err
	}

	src.iter.wav, err = getWav(src.wav, start, end, src.wav.FS)
	if err != nil {
		return err
	}
	if frameSize < 1 {
		src.iter.frameSize = len(src.iter.wav.Samples)
		src.iter.stepSize = src.iter.frameSize
	} else {
		src.iter.frameSize = frameSize
		src.iter.stepSize = stepSize
	}
	return nil
}

// Next loads the next available waveform into the source. Returns Done when all waveforms have been processed.
//func (src *SourceProc) Next() error {
//	return src.NextSegment(0, -1)
//}

// NextSegment loads a segment of the next available waveform into the source. Returns Done when all waveforms have been processed.
// See also Waveform.NextSegment() for details.
//func (src *SourceProc) NextSegment(start, end int) error {
func (src *SourceProc) Next() error {
	var err error
	//	src.wav, err = src.iter.NextSegment(start, end)
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

	// Undo rewind changes.
	src.iter.winType = src.winType
	src.iter.winData = src.winData

	return nil
}

// Get implements the dsp.Processer interface.
// If window option is used, window size must be less or equal than frameSize. If smaller, remaining samples are zero padded.
func (src *SourceProc) Get(idx uint32) (dsp.Value, error) {
	in, err := src.iter.Frame(int(idx))
	if err != nil {
		return nil, err
	}
	// No windowing, we are done.
	if src.iter.frameSize == 0 {
		return in, nil
	}
	inSize := in.Shape[0]
	if src.iter.frameSize > inSize {
		return nil, fmt.Errorf("window size [%d] is larger than input vector size [%d]", src.iter.frameSize, inSize)
	}
	v := narray.New(inSize)
	if src.iter.winType > 0 {
		for i, w := range src.iter.winData {
			v.Data[i] = in.Data[i] * w
		}
	} else {
		copy(v.Data, in.Data)
	}
	// Zero padding.
	for i := src.iter.frameSize; i < inSize; i++ {
		in.Data[i] = 0
	}
	return v, nil
}

// ID returns the id of the current waveform.
func (src *SourceProc) ID() string {
	return src.wav.ID
}

// NumFrames returns the number of frames in the current waveform.
func (src *SourceProc) NumFrames() int {
	return src.iter.NumFrames()
}

// Mean returns the mean of the waveform samples as they were read from the source.
func (src *SourceProc) Mean() float64 {
	return src.wav.Mean()
}

// SD returns the standard deviation of the waveform samples as they were read from the source.
func (src *SourceProc) SD() float64 {
	return src.wav.SD()
}

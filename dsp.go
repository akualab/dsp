// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsp

import "sync"

// Value is the type used to exhcange values between processors.
type Value []float64

// Copy creates a copy of a Value.
// Input values should be treated as read-only because
// they may be shared with other processors.
func (v Value) Copy() Value {
	vcopy := make(Value, len(v), len(v))
	copy(vcopy, v)
	return vcopy
}

// ToChan is the type used to send a Value to a channel.
type ToChan chan<- Value

// FromChan is the type used to receive a Value from a channel.
type FromChan <-chan Value

// The Processor interface must be implemented by all processors.
type Processor interface {
	RunProc(in []FromChan, out []ToChan) error
}

// The Resetter interface must be implemented by processors that have state that must
// be reset before processing a new stream.
type Resetter interface {
	Reset()
}

// ProcFunc is an adapter type that allows the use of ordinary
// functions as Processors.  If f is a function with the appropriate
// signature, FilterFunc(f) is a Processor that calls f.
type ProcFunc func([]FromChan, []ToChan) error

// RunProc calls this function. It implements the Processor interface.
func (f ProcFunc) RunProc(in []FromChan, out []ToChan) error { return f(in, out) }

func runProc(p Processor, in []FromChan, out []ToChan, e *procErrors) {
	e.record(p.RunProc(in, out))
	CloseOutputs(out)
}

// Sequence is a helper method that returns a processor that is the
// concatenation of all processor arguments.
// The output of a processor is fed as input to the next processor.
func (app *App) Sequence(procs ...Processor) Processor {
	if len(procs) == 1 {
		return procs[0]
	}
	return ProcFunc(func(in []FromChan, out []ToChan) error {
		input := in[0]
		for _, p := range procs {
			c := app.Wire()
			app.ConnectOne(p, c, input)
			input = c
		}
		for v := range input {
			SendValue(v, out)
		}
		return app.Error()
	})
}

// Run executes a sequence of processors.
// It returns either nil, an error if any processor reported an error.
func (app *App) Run(procs ...Processor) FromChan {
	p := app.Sequence(procs...)
	in := app.Wire()
	close(in)
	out := app.Wire()
	app.ConnectOne(p, out, in)

	return out
}

// App defines a DSP application.
type App struct {
	// App name.
	Name string
	// Default buffer size for connecting channels.
	BufferSize int
	e          *procErrors
}

// NewApp returns a new app.
func NewApp(name string, bufferSize int) *App {
	return &App{Name: name,
		e:          &procErrors{},
		BufferSize: bufferSize,
	}
}

// Connect connects multiple inputs and outputs to a processor.
func (app *App) Connect(p Processor, in []FromChan, out []ToChan) {
	go runProc(p, in, out, app.e)
}

// ConnectOne connects multiple inputs and one output to a processor.
func (app *App) ConnectOne(p Processor, out ToChan, ins ...FromChan) {
	//go runProc(p, PIO{Out: []ToChan{out}, In: ins}, app.e)
	go runProc(p, ins, []ToChan{out}, app.e)
}

// Error returns error if any.
func (app *App) Error() error {
	return app.e.getError()
}

// Wire creates a channel for wiring processors.
func (app *App) Wire() chan Value {
	return make(chan Value, app.BufferSize)
}

// CloseOutputs closes all the output channels.
func CloseOutputs(out []ToChan) {
	for _, o := range out {
		close(o)
	}
}

// SendValue sends a value to all the output channels.
func SendValue(v Value, out []ToChan) {

	for _, o := range out {
		o <- v
	}
}

// procErrors records errors accumulated during the execution of a processor.
type procErrors struct {
	mu  sync.Mutex
	err error
}

func (e *procErrors) record(err error) {
	if err != nil {
		e.mu.Lock()
		if e.err == nil {
			e.err = err
		}
		e.mu.Unlock()
	}
}

func (e *procErrors) getError() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.err
}

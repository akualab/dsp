// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package dsp provides processors that can be chained together to build
digital signal processing systems.

Processors are chained together using channels. Processors send data of
type Value which is a synonim of []float64. This is the only type allowed.

Data is pushed from a source and results can be read from any processor
in the chain.

Procesors have a single output and zero or more inputs.

The following example implements a simple pipeline where the output of a processor
is the single input to the next processor:

	app := NewApp("A simple pipeline.", 1000)

	out := app.Run(
        dsp.Source(64, 2).Use(NewSquare(1, 0, 4, 4)),
        dsp.Window(64).Use(Hamming),
        dsp.WriteValues(os.Stdout, true),
	)

The processed vectors can be read from the "out" channel. As we can see all
the wiring between processors is hidden.

A sequence of processors can be converted into a single composite processor as follows:

	newProc := app.Sequence(
        dsp.Source(64, 2).Use(NewSquare(1, 0, 4, 4)),
        dsp.Window(64).Use(Hamming),
        dsp.WriteValues(os.Stdout, true),
	)

This can be useful to hide some complexity when building an app.

A more realistic application will have processors with multiple
inputs and will generate multiple outputs.

For example:


Input values should be treated as read-only because they may be shared with other processors.
To create a copy, use newValue = inputValue.Copy().

CREDITS:

I adapted the design from https://github.com/ghemawat/stream by Sanjay Ghemawat.
Package stream is designed to process streams of text by chaining filters.

*/
package dsp

import "sync"

type Value []float64

// Creates a copy of teh value.
// Input values should be treated as read-only because
// they may be shared with other processors.
func (v Value) Copy() Value {
	vcopy := make(Value, len(v), len(v))
	copy(vcopy, v)
	return vcopy
}

type ToChan chan<- Value   // can only send to the channel
type FromChan <-chan Value // can only receive from the channel

type Processor interface {
	RunProc(PIO) error
}

// Processor IO (PIO) has the processor input and output channels.
// Slice PIO.In has channels that send values to the processor.
// Slice PIO.Out has channels that send values from the processor.
type PIO struct {
	In  []FromChan // []<-chan Value
	Out []ToChan   // chan<- Value
}

// Adds an output channel to processor.
func (pio *PIO) AddOut(out ToChan) {
	pio.Out = append(pio.Out, out)
}

// Adds an input channel to processor.
func (pio *PIO) AddIn(in FromChan) {
	pio.In = append(pio.In, in)
}

// ProcFunc is an adapter type that allows the use of ordinary
// functions as Processors.  If f is a function with the appropriate
// signature, FilterFunc(f) is a Processor that calls f.
type ProcFunc func(PIO) error

// RunProc calls this function. It implements the Processer interface.
func (f ProcFunc) RunProc(pio PIO) error { return f(pio) }

func runProc(p Processor, pio PIO, e *procErrors) {
	e.record(p.RunProc(pio))
	CloseOutputs(pio)
}

// Sequence returns a processor that is the concatenation of all processor arguments.
// The output of a processor is fed as input to the next processor.
func (app *App) Sequence(procs ...Processor) Processor {
	if len(procs) == 1 {
		return procs[0]
	}
	return ProcFunc(func(pio PIO) error {
		in := pio.In[0]
		for _, p := range procs {
			c := app.Wire()
			app.Connect(p, MakePIO(in, c))
			in = c
		}
		for v := range in {
			SendValue(v, pio)
		}
		return app.Error()
	})
}

// Run executes the sequence of processors.
// It returns either nil, an error if any filter reported an error.
func (app *App) Run(procs ...Processor) FromChan {
	p := app.Sequence(procs...)
	in := app.Wire()
	close(in)
	out := app.Wire()
	app.Connect(p, PIO{[]FromChan{in}, []ToChan{out}})

	return out
}

type App struct {
	// App name.
	Name string
	// Default buffer size for connection channels.
	BufferSize int
	e          *procErrors
}

// Returns a new app.
func NewApp(name string, bufferSize int) *App {

	return &App{Name: name,
		e:          &procErrors{},
		BufferSize: bufferSize,
	}
}

// Connects multiple inputs and outputs to a processor.
func (app *App) Connect(p Processor, pio PIO) {
	go runProc(p, PIO{Out: pio.Out, In: pio.In}, app.e)
}

// Connects multiple inputs and one output to a processor.
func (app *App) ConnectOne(p Processor, out ToChan, ins ...FromChan) {
	go runProc(p, PIO{Out: []ToChan{out}, In: ins}, app.e)
}

// Returns error if any.
func (app *App) Error() error {
	return app.e.getError()
}

// Creates a channel for wiring processors.
func (app *App) Wire() chan Value {
	return make(chan Value, app.BufferSize)
}

// Closes all the output channels.
func CloseOutputs(pio PIO) {
	for _, out := range pio.Out {
		close(out)
	}
}

// Sends a value to all the output channels.
func SendValue(v Value, pio PIO) {

	for _, out := range pio.Out {
		out <- v
	}
}

// Creates a PIO with a single input and output.
func MakePIO(in FromChan, out ToChan) PIO {
	return PIO{[]FromChan{in}, []ToChan{out}}
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

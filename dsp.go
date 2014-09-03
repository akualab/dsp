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


*/
package dsp

import "sync"

type Value []float64

type ToChan chan<- Value   // can only send to the channel
type FromChan <-chan Value // can only receive from the channel

type Processor interface {
	RunProc(Arg) error
}

// Arg contains the data passed to a Processor. Arg.In is a slice of channels that
// produce the inputs to the processor, and Arg.Out is a slice of channels that
// receive the outputs from the processor.
type Arg struct {
	In  []FromChan //[]<-chan Value
	Out ToChan     //chan<- Value
}

// ProcFunc is an adapter type that allows the use of ordinary
// functions as Processors.  If f is a function with the appropriate
// signature, FilterFunc(f) is a Processor that calls f.
type ProcFunc func(Arg) error

// RunProc calls this function. It implements the Processer interface.
func (f ProcFunc) RunProc(arg Arg) error { return f(arg) }

func runProc(p Processor, arg Arg, e *procErrors) {
	e.record(p.RunProc(arg))
	close(arg.Out)
}

// Sequence returns a processor that is the concatenation of all processor arguments.
// The output of a processor is fed as input to the next processor.
func (app *App) Sequence(procs ...Processor) Processor {
	if len(procs) == 1 {
		return procs[0]
	}
	return ProcFunc(func(arg Arg) error {
		in := arg.In[0]
		for _, p := range procs {
			c := app.Wire()
			app.Connect(p, c, in)
			in = c
		}
		for v := range in {
			arg.Out <- v
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
	app.Connect(p, out, in)

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

// Connects output and input channels to a processor and activates
// the processor.
func (app *App) Connect(p Processor, out ToChan, ins ...FromChan) {
	go runProc(p, Arg{Out: out, In: ins}, app.e)
}

// Returns error if any.
func (app *App) Error() error {
	return app.e.getError()
}

// Creates a channel for wiring processors.
func (app *App) Wire() chan Value {
	return make(chan Value, app.BufferSize)
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

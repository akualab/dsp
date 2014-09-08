// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package dsp provides processors that can be chained together to build
digital signal processing systems.

Digital signals are represented as sequence of numbers where each number
is associated with a disctrete time. Discrete time is represented as a
squence of integers that can correspond to physycal time sampled at fixed
time intervals.

All operations are synchronous because all values must correspond to a discrete
time. For example, given and operation such as x[n] = f[n-1] + g[n-2], the
value x[4] is calculated using delayed values f[3] and g[2].

Processors are chained together using channels. Processors send data of
type Value which is a synonim of []float64. In mathematical terms, all inputs
and outputs are vectors. The dimension of the vectors is set by the processors
using configuration parameters, teh dimension of the input vestors, or by any
other method.

Data is pushed from a source and results can be read from any processor
in the chain.

Procesors have zero or more outputs and zero or more inputs.

A processor with multiple outputs sends the send value to other processors.

A processor with multiple inputs receive values from other processors. The inputs
can be interchangeable (they all have the same behavior) or not internchangeable.

Processors are connected using buffered channels of type (chan Value). A full channel
will block incoming values. If the application graph is not properly built or the
output values are not read, the application will eventually deadlock.

While the operations are synchronous the underlying computation is asynchronous.
Processors will consume input data and write outputs as long as they are not blocked.
This approach makes it possible to do parallel processing in hosts with multiple
CPUs. Processor must be go routine safe.

The following example implements a simple pipeline where the output of a processor
is the single input to the next processor:

	app := NewApp("A simple pipeline.", 1000)

	out := app.Run(
        dsp.Source(64, 2).Use(NewSquare(1, 0, 4, 4)),
        dsp.Window(64).Use(Hamming),
        dsp.WriteValues(os.Stdout, true),
	)

The processed vectors can be read from the "out" channel. In linear apps, all
the wiring between processors is hidden.

A pipeline can easily be converted into a single composite processor:

	newProc := app.Sequence(
        dsp.Source(64, 2).Use(NewSquare(1, 0, 4, 4)),
        dsp.Window(64).Use(Hamming),
        dsp.WriteValues(os.Stdout, true),
	)

A more realistic application will have processors with multiple
inputs and multiple outputs. The Builder functions provide a set of
tools to create the application graph. See the examples.

Convention: Input values should be treated as read-only because
they may be shared with other processors.
To quickly create a copy, do newValue = inputValue.Copy().

CREDITS:

I adapted the design from https://github.com/ghemawat/stream by Sanjay Ghemawat.
Package stream is designed to process streams of text by chaining filters.

CONTACT:

Leo Neumeyer
leo@akualab.com

*/
package dsp

import (
	"fmt"
	"sync"
)

type Value []float64

// Creates a copy of the value.
// Input values should be treated as read-only because
// they may be shared with other processors.
func (v Value) Copy() Value {
	vcopy := make(Value, len(v), len(v))
	copy(vcopy, v)
	return vcopy
}

type ToChan chan<- Value   // can only send to the channel
type FromChan <-chan Value // can only receive from the channel

// Processors must implement this interface.
type Processor interface {
	RunProc(in In, out Out) error
}

// Input to processor.
type In struct {
	From []FromChan // []<-chan Value
}

// Output from processor.
type Out struct {
	To []ToChan // chan<- Value
}

// Returns a pair of initialized In and Out objects.
func NewIO() (In, Out) {
	return In{From: []FromChan{}}, Out{To: []ToChan{}}
}

// Adds an output channel to processor.
func (out *Out) Add(ch ToChan) {
	out.To = append(out.To, ch)
}

// Adds an input channel to processor.
func (in *In) Add(ch FromChan) {
	in.From = append(in.From, ch)
}

// Get input channel for index.
func (in *In) Get(idx int) (FromChan, error) {
	if idx < len(in.From) {
		return in.From[idx], nil

	}
	return nil, fmt.Errorf("there is no input with index [%d]", idx)
}

// ProcFunc is an adapter type that allows the use of ordinary
// functions as Processors.  If f is a function with the appropriate
// signature, FilterFunc(f) is a Processor that calls f.
type ProcFunc func(In, Out) error

// RunProc calls this function. It implements the Processor interface.
func (f ProcFunc) RunProc(in In, out Out) error { return f(in, out) }

func runProc(p Processor, in In, out Out, e *procErrors) {
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
	return ProcFunc(func(in In, out Out) error {
		input := in.From[0]
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

// A DSP application.
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
func (app *App) Connect(p Processor, in In, out Out) {
	go runProc(p, in, out, app.e)
}

// Connects multiple inputs and one output to a processor.
func (app *App) ConnectOne(p Processor, out ToChan, ins ...FromChan) {
	//go runProc(p, PIO{Out: []ToChan{out}, In: ins}, app.e)
	go runProc(p, In{From: ins}, Out{To: []ToChan{out}}, app.e)
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
func CloseOutputs(out Out) {
	for _, o := range out.To {
		close(o)
	}
}

// Sends a value to all the output channels.
func SendValue(v Value, out Out) {

	for _, o := range out.To {
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

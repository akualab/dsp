// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsp

import "sync"

type Value []float64

type Output chan<- Value // can only send to the channel
type Input <-chan Value  // can only receive from the channel

type Processor interface {
	RunProc(Arg) error
}

// Arg contains the data passed to a Processor. Arg.In is a slice of channels that
// produce the inputs to the processor, and Arg.Out is a slice of channels that
// receive the outputs from the processor.
type Arg struct {
	In  []Input //[]<-chan Value
	Out Output  //chan<- Value
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
	//	for _ = range arg.In { // Discard all unhandled input
	//	}
}

type App struct {
	Name       string
	BufferSize int
	e          *procErrors
}

func NewApp(name string, bufferSize int) *App {

	return &App{Name: name,
		e:          &procErrors{},
		BufferSize: bufferSize,
	}
}

func (app *App) Connect(p Processor, out Output, ins ...Input) {
	go runProc(p, Arg{Out: out, In: ins}, app.e)
}

func (app *App) Error() error {
	return app.e.getError()
}

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

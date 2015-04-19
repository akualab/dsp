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
type Value which is a synonym of []float64. In mathematical terms, all inputs
and outputs are vectors. The dimension of the vectors is set by the processors
using configuration parameters, the dimension of the input vectors, or by any
other method.

Data is pushed from a source and results can be read from any processor
in the chain.

Procesors have zero or more outputs and zero or more inputs.

A processor can have multiple outputs to sends values to multiple processors.

A processor with multiple inputs receive values from other processors. The inputs
can be interchangeable (all inputs are processed identically) or not internchangeable
(processor inputs have specialized behavior).

Processors are connected using buffered channels of type (chan Value). A full channel
will block incoming values. If the application graph is not properly built or the
output values are not read, the application will eventually deadlock.

While the operations are synchronous the underlying computation is asynchronous.
Processors will consume input data and write outputs as long as they are not blocked.
This approach makes it possible to do parallel processing in hosts with multiple
CPUs. Processors must be "go routine"-safe.

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
tools to create the application graph.

For a comrehensive example see examples/speech2/main.go.

Convention: Input values should be treated as read-only because
they may be shared with other processors.
To quickly create a copy, do newValue = inputValue.Copy().

CREDITS:

I adapted the design from https://github.com/ghemawat/stream by Sanjay Ghemawat.
Package stream is designed to process streams of text by chaining filters. Package stream
was itself adapted from Gustavo Niemeyer's "pipe" https://gopkg.in/pipe.v2 . These are two
interesting designs for piping data throuch channels. Check them out.

CONTACT:

Leo Neumeyer
leo@akualab.com

*/
package dsp

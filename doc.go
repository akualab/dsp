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

Tn application is a graph of processors where the input of a processor is read
from the output of another processor. Each data object moving out of a processor
is associated with a frame index. The data frame rate depends on the processors
which may interpolate or decimate the input data.

The architecture uses a "pull" approach. That is, when a frame is requested for
one of the processor's outputs, the computation chain propagates all the way to
the source.

Computed frames are cached by the processors to avoid recomputing the saem frame
more than once. This is important when a frame is requested multiple times by various
processors or by a single processor. (For example, to do a moving average, the same
input frames may be requested multiple times.)


A more realistic application will have processors with multiple
inputs and multiple outputs. The Builder functions provide a set of
tools to create the application graph.

For a comrehensive example see examples/speech2/main.go.

Convention: Input values should be treated as read-only because
they may be shared with other processors.
To quickly create a copy, do newValue = inputValue.Copy().

CONTACT:

Leo Neumeyer
leo@akualab.com

*/
package dsp

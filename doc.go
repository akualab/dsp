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

Sequences are divided into equally-spaced frames such that each frame corresponds
to a fixed number of samples.

Processors can return values for a frame index or a global value for the entire sequence.
To compute frame-level values, the processor must implement the Framer interface
which defines the "Get(int) (Value, error)" method. To return a global value the processor must
implement the OneValuer interface which defines the "Get() (Value, error.)" method.

The application is a graph of processors where the input of a processor is read
from the output of another processor. Processors exchange values of interface type Value. To perform
operations, the Processors must agree on the underlying types of Value.

Values are lazily computed when they are requested by a consumer. That is, when a value is requested for
one of the processor's outputs, the computation chain propagates all the way to
the source.

To achieve high performance, computed frames are cached by the processors. If the cache capacity is big enough,
values are only computed once. (For example, to do a moving average, the same input frames may be
requested multiple times.)

For a comrehensive example see examples/speech2/main.go.

Convention: Input values should be treated as read-only because
they may be shared with other processors.

CONTACT:
Leo Neumeyer
leo@akualab.com

*/
package dsp

package stream

import "sync"

type Value []float64

type Processor interface {
	RunProc(Arg) error
}

// Arg contains the data passed to a Processor. Arg.In is a slice of channels that
// produce the inputs to the processor, and Arg.Out is a slice of channels that
// receive the outputs from the processor.
type Arg struct {
	In  <-chan Value
	Out chan<- Value
}

// ProcFunc is an adapter type that allows the use of ordinary
// functions as Processors.  If f is a function with the appropriate
// signature, FilterFunc(f) is a Processor that calls f.
type ProcFunc func(Arg) error

// RunProc calls this function. It implements the Processer interface.
func (f ProcFunc) RunProc(arg Arg) error { return f(arg) }

const channelBuffer = 1000

func runProc(p Processor, arg Arg, e *procErrors) {
	e.record(p.RunProc(arg))
	close(arg.Out)
	for _ = range arg.In { // Discard all unhandled input
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

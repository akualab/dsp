// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsp

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

// Called a Proc that has no ProcFunc set.
var ErrNoFunc = errors.New("no ProcFunc set")

// Returned when frame index is out of bounds. Can be used as a termination flag.
var ErrOOB = errors.New("frame index out of bounds")

// Framer is the type used to exchange values between processors.
type Framer interface {
	// Create a copy of a frame. This is important because frames should never
	// be modified in place.
	Copy() Value
}

// The Processer interface must be implemented by all processors.
type Processer interface {
	Get(int) (Value, error)
	SetInputs(...Processer)
	Reset()
}

// ProcFunc is the type used to implement processing functions.
type ProcFunc func(int, ...Processer) (Value, error)

// Proc can be embedded in objects that implement the Processer interface.
type Proc struct {
	f      ProcFunc
	inputs []Processer
	cache  *cache
}

// NewProc creates a new Proc.
func NewProc(bufSize int, f ProcFunc) *Proc {
	return &Proc{
		f:     f,
		cache: newCache(bufSize),
	}
}

// SetInputs sets the inputs for a processor.
func (bp *Proc) SetInputs(inputs ...Processer) {
	bp.inputs = inputs
}

// Reset - override this method to reset the processor state.
func (bp *Proc) Reset() {
	bp.cache.clear()
}

// Get - returns value for index.
func (bp *Proc) Get(idx int) (Value, error) {
	if idx < 0 {
		return nil, ErrOOB
	}
	val, ok := bp.cache.get(idx)
	if ok {
		return val, nil
	}
	if bp.f != nil {
		v, e := bp.f(idx, bp.inputs...)
		if e != nil {
			return nil, e
		}
		bp.cache.set(idx, v)
		return v, nil
	}
	return nil, ErrNoFunc
}

// SetCache sets the value in the cache.
func (bp *Proc) SetCache(idx int, val Value) {
	bp.cache.set(idx, val)
}

// GetCache gets value from cache.
func (bp *Proc) GetCache(idx int) (Value, bool) {
	val, ok := bp.cache.get(idx)
	return val, ok
}

// ClearCache clears the cache.
func (bp *Proc) ClearCache() {
	bp.cache.clear()
}

// Inputs returns the input processors.
func (bp *Proc) Inputs() []Processer {
	return bp.inputs
}

// Input returns one of the processor inputs.
func (bp *Proc) Input(n int) Processer {
	return bp.inputs[n]
}

// App defines a DSP application.
type App struct {
	// App name.
	Name   string
	procs  map[string]Node
	inputs map[Node][]Node
}

// Node is a node in the processor graph.
type Node struct {
	name string
	typ  Processer
}

// Name of the processor.
func (n Node) Name() string {
	return n.name
}

// Get returns the value for frame index.
func (n Node) Get(idx int) (Value, error) {
	return n.typ.Get(idx)
}

// NewApp returns a new app.
func NewApp(name string) *App {
	return &App{
		Name:   name,
		procs:  make(map[string]Node),
		inputs: make(map[Node][]Node),
	}
}

// NodeByName returns a node from the processor graph.
func (app *App) NodeByName(name string) Node {
	proc, ok := app.procs[name]
	if !ok {
		panic(fmt.Errorf("no processor named [%s] in builder graph", name))
	}
	return proc
}

// NodesByName converts a list of names into a list of nodes.
func (app *App) NodesByName(names ...string) ([]Node, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("the names argument is empty, need at least one name")
	}
	nodes := []Node{}
	for _, name := range names {
		node, ok := app.procs[name]
		if !ok {
			return nil, fmt.Errorf("no processor named [%s] in builder graph", name)
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// Add adds a processor with a name.
func (app *App) Add(name string, p Processer) Node {
	nodeName := strings.TrimSpace(name)
	_, ok := app.procs[nodeName]
	if ok {
		panic(fmt.Errorf("there is already a processor named [%s] in builder graph", nodeName))
	}
	n := Node{name: nodeName, typ: p}
	app.procs[nodeName] = n
	return n
}

// Connect connects processor inputs. Example:
//    var y,x1,x2 dsp.Node
//    ...
//    out := app.Connect(y, x1, x2)
// the output values of processors x1 and x2 are
// inputs to processor y. Returns node corresponding to processor y.
func (app *App) Connect(to Node, from ...Node) Node {
	inputs := []Processer{}
	for _, in := range from {
		inputs = append(inputs, in.typ)
	}
	to.typ.SetInputs(inputs...)
	app.inputs[to] = from
	return to
}

// Chain connects a sequence of processors
// as follows:
//    var p0, p1, p2, p3 dsp.Node
//    ...
//    out := app.Pipe(p0, p1, p2, p3)
// p0 <= p1 <= p2 <= p3 (the last processor in the chain is p0
// which is return by the method.
func (app *App) Chain(nodes ...Node) Node {
	for k := 0; k < len(nodes)-1; k++ {
		nodes[k].typ.SetInputs(nodes[k+1].typ)
		app.inputs[nodes[k]] = []Node{nodes[k+1]}
	}
	return nodes[0]
}

// Reset resets all processors in preparation for a new stream.
func (app *App) Reset() {
	for _, p := range app.procs {
		p.typ.Reset()
	}
}

func (app *App) String() string {

	var buf bytes.Buffer
	for name, node := range app.procs {
		buf.WriteString(fmt.Sprintf("proc: %s, inputs: | ", name))
		for _, in := range app.inputs[node] {
			buf.WriteString(fmt.Sprintf("%s | ", in))
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

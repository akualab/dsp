// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsp

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Called a Proc that has no ProcFunc set.
var ErrNoFunc = errors.New("no ProcFunc set")

// Returned when frame index is out of bounds. Can be used as a termination flag.
var ErrOOB = errors.New("frame index out of bounds")

// The Processer interface is the common pnterface for all processors.
type Processer interface {
}

// The Framer interface is used to get a frame values.
type Framer interface {
	Get(int) (Value, error)
}

// The OneValuer interface is used to get a single value from a stream.
type OneValuer interface {
	Get() (Value, error)
}

// The Inputter interface is used to set processor inputs.
type Inputter interface {
	SetInputs(...Processer)
}

// The Resetter interface is used to reset the state of a stateful processor.
type Resetter interface {
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

// Framer returns processor input #n as a Framer type.
func (bp *Proc) Framer(n int) Framer {
	return bp.inputs[n].(Framer)
}

// OneValuer returns processor input #n as a OneValuer type.
func (bp *Proc) OneValuer(n int) OneValuer {
	return bp.inputs[n].(OneValuer)
}

// OneProcFunc is the type used to implement processing functions that return a single value per stream.
type OneProcFunc func(...Processer) (Value, error)

// OneProc can be embedded in structs that need to implement the OneValuer interface.
type OneProc struct {
	f      OneProcFunc
	inputs []Processer
	cache  Value
}

// NewOneProc creates a new Proc.
func NewOneProc(f OneProcFunc) *OneProc {
	return &OneProc{f: f}
}

// SetInputs sets the inputs for a processor.
func (bp *OneProc) SetInputs(inputs ...Processer) {
	bp.inputs = inputs
}

// Reset - override this method to reset the processor state.
func (bp *OneProc) Reset() {
	bp.cache = Value(nil)
}

// Get - returns one value for stream.
func (bp *OneProc) Get() (Value, error) {
	if bp.cache != nil {
		return bp.cache, nil
	}
	if bp.f != nil {
		v, e := bp.f(bp.inputs...)
		if e != nil {
			return nil, e
		}
		bp.cache = v
		return v, nil
	}
	return nil, ErrNoFunc
}

// Inputs returns the input processors.
func (bp *OneProc) Inputs() []Processer {
	return bp.inputs
}

// Framer returns processor input #n as a Framer type.
func (bp *OneProc) Framer(n int) Framer {
	return bp.inputs[n].(Framer)
}

// OneValuer returns processor input #n as a OneValuer type.
func (bp *OneProc) OneValuer(n int) OneValuer {
	return bp.inputs[n].(OneValuer)
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

// Proc returns the processor associated to this node.
func (n Node) Proc(idx int) Processer {
	return n.typ
}

// Get returns value for frame. Underlying processor must implement the Framer interface.
func (n Node) Get(idx int) (Value, error) {
	return n.typ.(Framer).Get(idx)
}

// GetOne returns value for frame.
func (n Node) GetOne() (Value, error) {
	return n.typ.(OneValuer).Get()
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
	inputter, ok := to.typ.(Inputter)
	if !ok {
		panic("tried to set inputs on a processor that does not implement the Inputter interface")
	}
	inputs := []Processer{}
	for _, in := range from {
		inputs = append(inputs, in.typ)
	}
	inputter.SetInputs(inputs...)
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
		inputter, ok := nodes[k].typ.(Inputter)
		if !ok {
			panic("tried to set inputs on a processor that does not implement the Inputter interface")
		}
		inputter.SetInputs(nodes[k+1].typ)
		app.inputs[nodes[k]] = []Node{nodes[k+1]}
	}
	return nodes[0]
}

// Reset resets processors that implement the Resetter interface.
// Should be called in preparation for a new stream when processors have state.
func (app *App) Reset() {
	for _, node := range app.procs {
		res, ok := node.typ.(Resetter)
		if ok {
			res.Reset()
		}
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

// Get is a convenience function to get the value of a processor.
// The underlying type of p may be a Framer or a OneValuer. If the type
// is a OneValuer, the idx argument is ignored.
func Get(p Processer, idx int) (val Value, err error) {
	switch input := p.(type) {
	case Framer:
		val, err = input.Get(idx)
		if err != nil {
			return nil, err
		}
	case OneValuer:
		val, err = input.Get()
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported input type %s", reflect.TypeOf(p))
	}
	return val, nil
}

// Processers is a slice of Processer values.
type Processers []Processer

// Get is a convenience method to get the value of the input at index 0.
func (ps Processers) Get(idx int) (Value, error) {
	framer, ok := ps[0].(Framer)
	if !ok {
		return nil, fmt.Errorf("processor does not implement the Framer interface")
	}
	return framer.Get(idx)
}

// Framer is a convenience method that checks that input #n exist and implements the Framer interface.
func (ps Processers) Framer(n int) (Framer, error) {
	if len(ps) < n {
		return nil, fmt.Errorf("processer slice has size %d, expected at least %d inputs", len(ps), n)
	}
	framer, ok := ps[n].(Framer)
	if !ok {
		return nil, fmt.Errorf("processer #%d does not implement the Framer interface", n)
	}
	return framer, nil
}

// OneValuer is a convenience method that checks that input #n exist and implements the OneValuer interface.
func (ps Processers) OneValuer(n int) (OneValuer, error) {
	if len(ps) < n {
		return nil, fmt.Errorf("processer slice has size %d, expected at least %d inputs", len(ps), n)
	}
	ov, ok := ps[n].(OneValuer)
	if !ok {
		return nil, fmt.Errorf("processer #%d does not implement the OneValuer interface", n)
	}
	return ov, nil
}

// CheckInputs is a convenience method that checks that the first n inputs exist and implement the Framer interface.
func (ps Processers) CheckInputs(n int) ([]Framer, error) {
	if len(ps) < n {
		return nil, fmt.Errorf("processer slice has size %d, expected at least %d inputs", len(ps), n)
	}

	framers := []Framer{}
	for i := 0; i < n; i++ {
		framer, ok := ps[i].(Framer)
		if !ok {
			return nil, fmt.Errorf("processer #%d does not implement the framer interface", i)
		}
		framers = append(framers, framer)
	}
	return framers, nil
}

// IsFramer returns true if the processor implements the Framer interface.
func IsFramer(p Processer) bool {
	_, ok := p.(Framer)
	return ok
}

// IsOneValuer returns true if the processor implements the OneValuer interface.
func IsOneValuer(p Processer) bool {
	_, ok := p.(OneValuer)
	return ok
}

// IsInputter returns true if the processor implements the Inputter interface.
func IsInputter(p Processer) bool {
	_, ok := p.(Inputter)
	return ok
}

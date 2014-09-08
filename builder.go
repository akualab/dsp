// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsp

import (
	"fmt"
	"log"

	"github.com/gonum/graph"
	g "github.com/gonum/graph/concrete"
)

// Helper to build complex application graphs.
type Builder struct {
	App      *App
	g        *g.DirectedGraph
	nodes    map[string]*node
	nodeByID map[graph.Node]*node
}

type node struct {
	n         graph.Node
	name      string
	proc      Processor
	fromChans []chan Value   //[]FromChan
	toChans   []chan Value   //[]ToChan
	inputIdx  map[string]int // edge to input index
}

func (app *App) NewBuilder() *Builder {

	b := &Builder{
		App:      app,
		g:        g.NewDirectedGraph(),
		nodes:    make(map[string]*node),
		nodeByID: make(map[graph.Node]*node),
	}

	return b
}

// Adds a processor with a name.
func (b *Builder) Add(name string, p Processor) {
	n := b.g.NewNode()
	b.nodes[name] = &node{
		n:         n,
		name:      name,
		proc:      p,
		toChans:   [](chan Value){},
		fromChans: [](chan Value){},
		inputIdx:  make(map[string]int),
	}
	b.nodeByID[n] = b.nodes[name]
}

// Add output node.
func (b *Builder) AddEndNode(name string) {
	b.Add(name, nil)
}

func (b *Builder) EndNodeChan(name string) (chan Value, error) {

	_, ok := b.nodes[name]
	if !ok {
		return nil, fmt.Errorf("there is no processor named %s", name)
	}

	if b.nodes[name].proc != nil {
		return nil, fmt.Errorf("processor %s is not an end node", name)
	}

	if len(b.nodes[name].toChans) == 0 {
		return nil, fmt.Errorf("processor %s has no output channel, make sure there is a connection to this end node", name)
	}
	return b.nodes[name].toChans[0], nil
}

// Adds a one way channel between two processors by name.
func (b *Builder) Connect(from, to string) error {
	return b.ConnectOrdered(from, to, -1)
}

// Adds a one way channel between two processors by name.
// The input of the receiver node is specified with an index.
// The index must be in the range {0...N-1} where N is the number
// of inputs.
// Use this method when the processors has multiple inputs that
// are not interchangeable. Otherwise, use Connect() instead.
func (b *Builder) ConnectOrdered(from, to string, idx int) error {

	var ok bool

	_, ok = b.nodes[from]
	if !ok {
		return fmt.Errorf("there is no processor named %s", from)
	}

	_, ok = b.nodes[to]
	if !ok {
		return fmt.Errorf("there is no processor named %s", to)
	}

	edge := g.Edge{T: b.nodes[from].n, H: b.nodes[to].n}
	b.g.AddDirectedEdge(edge, 1.0)
	imap := b.nodes[to].inputIdx
	k := fmt.Sprintf("%s-%s", edge.Tail(), edge.Head())
	if idx >= 0 {
		imap[k] = idx
		return nil
	}
	imap[k] = len(imap) // auto-increment.
	return nil
}

func (b *Builder) Run() {

	// Create one channel per edge.
	for _, e := range b.g.EdgeList() {
		edge := e.(g.WeightedEdge)
		from := b.nodeByID[edge.Tail()]
		to := b.nodeByID[edge.Head()]
		ch := make(chan Value, b.App.BufferSize)

		// We need to connect to a specific index in the input slice.
		k := fmt.Sprintf("%s-%s", edge.Tail(), edge.Head())
		idx := to.inputIdx[k] // the target input.
		s := &to.toChans      // slice of input channels
		if idx < len(*s) {
			(*s)[idx] = ch // insert and done.
		} else { // slice too short, augment.
			for i := len(*s); i < idx; i++ {
				*s = append(*s, nil) // fill out slice
			}
			*s = append(*s, ch) // insert as last element
		}

		from.fromChans = append(from.fromChans, ch)
		log.Printf("add channel from [%s] to [%s]", from.name, to.name)
	}

	// For each node, launch the processor.
	for _, nn := range b.g.NodeList() {
		node := b.nodeByID[nn]
		in, out := NewIO()
		for _, ch := range node.toChans {
			if ch == nil {
				fmt.Errorf("found a nil input channel in node %s - check the node connections", node.name)
			}
			in.Add(ch)
		}
		for _, ch := range node.fromChans {
			if ch == nil {
				fmt.Errorf("found a nil output channel in node %s - this should not happen, report teh bug", node.name)
			}
			out.Add(ch)
		}
		log.Printf("preparing to launch node [%s], in: %d, out: %d",
			node.name,
			len(node.toChans),
			len(node.fromChans))
		// launch! (skip end nodes)
		if node.proc != nil {
			go runProc(node.proc, in, out, b.App.e)
		}
	}
}

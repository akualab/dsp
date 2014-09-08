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
	fromChans []chan Value //[]FromChan
	toChans   []chan Value //[]ToChan
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
		n:    n,
		name: name,
		proc: p,
		//		toChans:   []ToChan{},
		toChans: [](chan Value){},
		//		fromChans: []FromChan{},
		fromChans: [](chan Value){},
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

	return nil
}

func (b *Builder) Run() {

	// Create one channel per edge.
	for _, e := range b.g.EdgeList() {
		edge := e.(g.WeightedEdge)
		from := b.nodeByID[edge.Tail()]
		to := b.nodeByID[edge.Head()]
		ch := make(chan Value, b.App.BufferSize)
		to.toChans = append(to.toChans, ch)
		from.fromChans = append(from.fromChans, ch)
		log.Printf("add channel from [%s] to [%s]", from.name, to.name)
	}

	// For each node, launch the processor.
	for _, nn := range b.g.NodeList() {
		node := b.nodeByID[nn]
		in, out := NewIO()
		for _, ch := range node.toChans {
			in.Add(ch)
		}
		for _, ch := range node.fromChans {
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

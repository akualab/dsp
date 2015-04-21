// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsp

import (
	"bytes"
	"fmt"
	"log"

	"github.com/gonum/graph"
	g "github.com/gonum/graph/concrete"
)

const tapSuffix = " tap"

// Builder is a helper to build complex application graphs.
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

// NewBuilder returns a new builder.
func (app *App) NewBuilder() *Builder {

	b := &Builder{
		App:      app,
		g:        g.NewDirectedGraph(),
		nodes:    make(map[string]*node),
		nodeByID: make(map[graph.Node]*node),
	}

	return b
}

// Add adds a processor with a name.
func (b *Builder) Add(name string, p Processor) string {
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
	return name
}

// Tap adds an output node with no processor.
// An output channel will be available from node "name".
func (b *Builder) Tap(name string) {
	tapNodeName := fmt.Sprintf("%s%s", name, tapSuffix)
	b.Add(tapNodeName, nil)
	b.Connect(name, tapNodeName)
}

// TapChan returns output channel from node.
func (b *Builder) TapChan(name string) chan Value {

	tapNodeName := fmt.Sprintf("%s%s", name, tapSuffix)

	_, ok := b.nodes[tapNodeName]
	if !ok {
		panic(fmt.Errorf("there is no end node conected from %s - use Tap(\"%s\") to create an end node", name, name))
	}
	if b.nodes[tapNodeName].proc != nil {
		panic(fmt.Errorf("end node [%s] has a non-nil processor, this should not happen, report bug.", tapNodeName))
	}
	if len(b.nodes[tapNodeName].toChans) == 0 {
		panic(fmt.Errorf("end node [%s] has no output channel, this should not happen, report bug", tapNodeName))
	}
	return b.nodes[tapNodeName].toChans[0]
}

// Connect adds a one way channel between two processors by name.
// Will panic if it finds an error.
func (b *Builder) Connect(from, to string) {
	b.ConnectOrdered(from, to, -1)
}

// ConnectOrdered adds a one way channel between two processors by name.
// The input of the receiver node is specified with an index.
// The index must be in the range {0...N-1} where N is the number
// of inputs.
// Use this method when the processors has multiple inputs that
// are not interchangeable. Otherwise, use Connect() instead.
// Will panic if it finds an error.
func (b *Builder) ConnectOrdered(from, to string, idx int) {

	var ok bool
	_, ok = b.nodes[from]
	if !ok {
		panic(fmt.Errorf("there is no processor named [%s]", from))
	}
	_, ok = b.nodes[to]
	if !ok {
		panic(fmt.Errorf("there is no processor named [%s]", to))
	}
	edge := g.Edge{T: b.nodes[from].n, H: b.nodes[to].n}
	b.g.AddDirectedEdge(edge, 1.0)
	imap := b.nodes[to].inputIdx
	k := fmt.Sprintf("%s-%s", edge.Tail(), edge.Head())
	if idx >= 0 {
		imap[k] = idx
		return
	}
	imap[k] = len(imap) // auto-increment.
}

// Run creates channels according to the graph specification and
// activates the processors.
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
		in, out := []FromChan{}, []ToChan{}
		for _, ch := range node.toChans {
			if ch == nil {
				fmt.Errorf("found a nil input channel in node [%s] - check the node connections", node.name)
			}
			in = append(in, ch)
		}
		for _, ch := range node.fromChans {
			if ch == nil {
				fmt.Errorf("found a nil output channel in node [%s] - this should not happen, report the bug", node.name)
			}
			out = append(out, ch)
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

func (b *Builder) String() string {

	var buf bytes.Buffer
	for name, node := range b.nodes {
		buf.WriteString(fmt.Sprintf("name: %s, ", name))
		for i, in := range node.fromChans {
			buf.WriteString(fmt.Sprintf("len(from[%d]): %d, ", i, len(in)))
		}
		for j, out := range node.toChans {
			buf.WriteString(fmt.Sprintf("len(to[%d]): %d, ", j, len(out)))
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

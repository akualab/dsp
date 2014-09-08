package dsp

import (
	"fmt"
	"strings"
)

// Read numbers from a string using two readers.
// The reader outpu vectors of length 4.
// Add the vectors from readers and multiply by 1.5.
func ExampleBuilder() {

	const input = "0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15"
	r1 := strings.NewReader(input)
	r2 := strings.NewReader(input)

	app := NewApp("Example Builder", 10)

	// Get the processsors.
	p1 := Reader(r1, NewReader(4))
	p2 := Reader(r2, NewReader(4))
	p3 := AddScaled(4, 1.5)

	// Use builder to greate the application graph.
	b := app.NewBuilder()

	// The nodes.
	b.Add("reader 1", p1)
	b.Add("reader 2", p2)
	b.Add("combo", p3)
	// The end nodes are special, don't have a processor.
	// We use them to get the output channels.
	b.AddEndNode("end")

	// The connections.
	// (Ommiting error checking for clarity.)
	b.ConnectOrdered("reader 1", "combo", 1)
	b.ConnectOrdered("reader 2", "combo", 0)
	b.Connect("combo", "end")

	// Run the app.
	b.Run()

	ch, _ := b.EndNodeChan("end")
	v := <-ch
	fmt.Printf("in     = [%s]\n", input)
	fmt.Println("out[0] =", v)
	v = <-ch
	fmt.Println("out[1] =", v)
	v = <-ch
	fmt.Println("out[2] =", v)
	// Output:
	// in     = [0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15]
	// out[0] = [0 3 6 9]
	// out[1] = [12 15 18 21]
	// out[2] = [24 27 30 33]
}

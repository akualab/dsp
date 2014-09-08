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

	// Create an app first. Give it a name and a default buffer size for the channels.
	app := NewApp("Example Builder", 10)

	// Use builder to create the application graph.
	b := app.NewBuilder()

	// App graph nodes with unique names.
	b.Add("reader 1", Reader(r1, NewReader(4)))
	b.Add("reader 2", Reader(r2, NewReader(4)))
	b.Add("combo", AddScaled(4, 1.5))

	// If you need an output channel from a node, use Tap().
	// Behind the scenes will create a channel connected to
	// a dummy end node with no processor.
	b.Tap("combo")

	// The connections. (App graph edges.)
	// ConnectOrdered() is only necessary when a processor
	// has multiple inputs that are not interchangeable.
	// AddScaled() has multiple inputs that are interchangeable.
	// We could have used Connect() instead of ConnectOrdered().
	// Will panic if there is a typo in a node name.
	b.ConnectOrdered("reader 1", "combo", 1)
	b.ConnectOrdered("reader 2", "combo", 0)

	// Run the app.
	b.Run()

	// Get the output channel.
	// Must be done after the app started.
	ch := b.TapChan("combo")
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

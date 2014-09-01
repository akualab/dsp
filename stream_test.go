package stream

import (
	"os"
	"testing"

	"fmt"
)

func TestBasic(t *testing.T) {

	e := &procErrors{}
	p1 := Random(4, 10)

	in := make(chan Value)
	close(in)
	c1 := make(chan Value, channelBuffer)
	go runProc(p1, Arg{In: in, Out: c1}, e)

	p2 := WriteValues(os.Stdout)
	c2 := make(chan Value, channelBuffer)
	go runProc(p2, Arg{In: c1, Out: c2}, e)

	for v := range c2 {
		_ = v
	}

	fmt.Println("error:", e.getError())
}

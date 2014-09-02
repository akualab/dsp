package stream

import (
	"os"
	"testing"

	"fmt"
)

func TestBasic(t *testing.T) {

	app := NewApp("Test", 1000)

	p1 := Source(4, 10).Use(NewNormal(88, 10, 2))

	w1 := app.MakeWire()
	app.Connect(p1, w1)

	p2 := WriteValues(os.Stdout)
	w2 := app.MakeWire()
	app.Connect(p2, w2, w1)

	for v := range w2 {
		_ = v
	}

	fmt.Println("error:", app.Error())
}

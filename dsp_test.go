package dsp

import (
	"testing"

	narray "github.com/akualab/narray/na64"
)

func numbers(idx uint32, in ...Processer) (Value, error) {
	return narray.NewArray([]float64{float64(idx)}, 1), nil
}

func square(idx uint32, in ...Processer) (Value, error) {
	v, err := in[0].Get(idx)
	if err != nil {
		return nil, err
	}
	x := []float64{v.Data[0] * v.Data[0]}
	return narray.NewArray(x, 1), nil
}

func TestGraph(t *testing.T) {

	app := NewApp("test")
	app.Add("numbers", NewProc(10, numbers))
	app.Add("square", NewProc(10, square))
	app.Connect("square", "numbers")
	sq := app.NewTap("square")

	var i uint32
	for ; i < 10; i++ {
		v, err := sq.Get(i)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(i, v)
	}
}

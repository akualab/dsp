package dsp

import "testing"

func numbers(idx uint32, in ...Processer) (Value, error) {
	return Value{float64(idx)}, nil
}

func square(idx uint32, in ...Processer) (Value, error) {
	v, err := in[0].Get(idx)
	if err != nil {
		return nil, err
	}
	return Value{v[0] * v[0]}, nil
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

package dsp

import "testing"

type TVal []float64

func (v TVal) Copy() Value {
	n := len(v)
	v2 := make(TVal, n, n)
	copy(v2, v)
	return v2
}

//func numbers(idx int, in ...Processer) (Value, error) {
func numbers(idx int, in ...Processer) (Value, error) {
	return TVal{float64(idx)}, nil
}

func square(idx int, in ...Processer) (Value, error) {
	v, err := Processers(in).Get(idx)
	if err != nil {
		return nil, err
	}
	val := v.(TVal)[0]
	return TVal{val * val}, nil
}

func TestGraph(t *testing.T) {

	app := NewApp("test")
	numbers := app.Add("numbers", NewProc(10, numbers))
	sq := app.Add("square", NewProc(10, square))
	app.Connect(sq, numbers)

	for i := 0; i < 10; i++ {
		v, err := sq.Get(i)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(i, v)
	}
}

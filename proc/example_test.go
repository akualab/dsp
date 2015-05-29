package proc_test

import (
	"fmt"

	"github.com/akualab/dsp"
	"github.com/akualab/dsp/proc"

	narray "github.com/akualab/narray/na64"
)

// Calculate Fibonacci values.
func ExampleApp_fibonacci() {
	app := dsp.NewApp("Fibonacci")
	fibo := app.Add("Fibo", Fibo(11))

	// Self loop.
	app.Connect(fibo, fibo)

	f10, e := fibo.Get(10)
	if e != nil {
		panic(e)
	}
	fmt.Println(f10.(*narray.NArray).Data[0])
	// Output:
	// 89
}

// Subtract a value "fm" from a Fibonacci series so the sum adds up to zero.
func ExampleApp_zmFibonacci() {
	app := dsp.NewApp("Fibonacci")
	fibo := app.Add("Fibo", Fibo(11))

	// Computes the mean of the series.
	fiboMean := app.Add("Fibo Mean", proc.Mean())

	// Subtracts mean from fibo values.
	fiboZM := app.Add("ZM Fibo", proc.Sub())

	// Make connections. Note that fiboMean is of tyep OneValuer and fibo is a Framer.
	// The reulting fiboZM is of type Framer. Values are computed only once and saved
	// in the processor cache.
	app.Connect(fibo, fibo)
	app.Connect(fiboMean, fibo)
	app.Connect(fiboZM, fibo, fiboMean)

	f10, e := fiboZM.Get(10)
	if e != nil {
		panic(e)
	}
	fmt.Println(f10.(*narray.NArray).Data[0])
	// Output:
	// 67.9090909090909
}

// Fibo is a processor that returns the Fibonacci value for index.
// N is the length of the series.
func Fibo(N int) dsp.Processer {
	return dsp.NewProc(20, func(idx int, in ...dsp.Processer) (dsp.Value, error) {
		if idx == 0 || idx == 1 {
			na := narray.New(1)
			na.Set(1, 0)
			return na, nil
		}
		if idx < 0 || idx >= N {
			return nil, dsp.ErrOOB
		}
		vm1, err := dsp.Processers(in).Get(idx - 1)
		if err != nil {
			return nil, err
		}
		vm2, err := dsp.Processers(in).Get(idx - 2)
		if err != nil {
			return nil, err
		}

		res := narray.Add(nil, vm2.(*narray.NArray), vm1.(*narray.NArray))
		return res, nil
	})
}

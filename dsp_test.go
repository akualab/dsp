package dsp

import (
	"math"
	"os"
	"testing"

	narray "github.com/akualab/narray/na64"
)

var (
	egy, fb, logfb []float64
	app            *App
	nf             int // num frames
)

func TestMain(m *testing.M) {

	// generate data.
	data := []float64{0.5, 1.0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	N := len(data)
	nf = 10      // numFrames
	ns := nf * N // num samples
	x := make([]float64, ns, ns)
	for i := 0; i < nf; i++ {
		copy(x[i*N:((i+1)*N)], data)
	}

	// Expected DFT Egy for one frame.
	y := make([]float64, N, N)
	copy(y, data)
	RealFT(y, N, true)
	egy = DFTEnergy(y)

	fb = []float64{}
	logfb = []float64{}
	for k, v := range testFBCoeff {
		sum := 0.0
		start := testFBInd[k]
		for i, w := range v {
			sum += w * egy[start+i]
		}
		fb = append(fb, sum)
		logfb = append(logfb, math.Log(sum))
	}

	// Use processors.
	app = NewApp("Test DFT Egy")
	app.Add("wav", wavSP(x))
	app.Add("spectrum", SpectralEnergy(3))
	app.Connect("spectrum", "wav")

	app.Add("filterbank", Filterbank(testFBInd, testFBCoeff))
	app.Add("log filterbank", Log())
	app.Connect("filterbank", "spectrum")
	app.Connect("log filterbank", "filterbank")

	os.Exit(m.Run())
}

func numbers(idx int, in ...Processer) (Value, error) {
	return narray.NewArray([]float64{float64(idx)}, 1), nil
}

func square(idx int, in ...Processer) (Value, error) {
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

	for i := 0; i < 10; i++ {
		v, err := sq.Get(i)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(i, v)
	}
}

func TestDFTEgyFeature(t *testing.T) {

	app.Reset()
	sp := app.NewTap("spectrum")

	// Compare results.
	cnt := 0
	for i := 0; i < nf; i++ {
		value, e := sp.Get(i)
		if e != nil {
			t.Fatal(e)
		}
		f := value.Data

		//		t.Log(f.Data)
		for k, v := range egy {
			if v != f[k] {
				t.Fatalf("mismatch for frame %d, elem %d - want %f, got %f", i, k, v, f[k])
			}
			cnt++
		}
	}
	t.Logf("compared %d values", cnt)
}

func TestFB(t *testing.T) {

	app.Reset()
	feat := app.NewTap("filterbank")
	t.Log("fb", fb)
	// Compare results.
	cnt := 0
	for i := 0; i < nf; i++ {
		value, e := feat.Get(i)
		if e != nil {
			t.Fatal(e)
		}
		f := value.Data

		//		t.Log(f.Data)
		for k, v := range fb {
			if v != f[k] {
				t.Fatalf("mismatch for frame %d, elem %d - want %f, got %f", i, k, v, f[k])
			}
			cnt++
		}
	}
	t.Logf("compared %d values", cnt)
}

func TestLogFB(t *testing.T) {

	app.Reset()
	feat := app.NewTap("log filterbank")
	t.Log("logfb", logfb)

	// Compare results.
	cnt := 0
	for i := 0; i < nf; i++ {
		value, e := feat.Get(i)
		if e != nil {
			t.Fatal(e)
		}
		f := value.Data

		//		t.Log(f.Data)
		for k, v := range logfb {
			if v != f[k] {
				t.Fatalf("mismatch for frame %d, elem %d - want %f, got %f", i, k, v, f[k])
			}
			cnt++
		}
	}
	t.Logf("compared %d values", cnt)
}

func wavSP(w []float64) Processer {
	return NewProc(1, func(idx int, in ...Processer) (Value, error) {
		return narray.NewArray(w, len(w)), nil
	})
}

var (
	testFBInd = []int{1, 3, 6}

	testFBCoeff = [][]float64{
		[]float64{1, 1, 1, 1},
		[]float64{1},
		[]float64{1, 1},
	}
)

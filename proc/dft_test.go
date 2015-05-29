package proc

import (
	"math"
	"os"
	"testing"

	"github.com/akualab/dsp"
	narray "github.com/akualab/narray/na64"
)

var (
	egy, fb, logfb []float64
	app            *dsp.App
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
	app = dsp.NewApp("Test DFT Egy")
	app.Connect(
		app.Add("spectrum", SpectralEnergy(3)),
		app.Add("wav", wavSP(x)),
	)
	app.Connect(
		app.Add("filterbank", Filterbank(testFBInd, testFBCoeff)),
		app.NodeByName("spectrum"),
	)
	app.Connect(
		app.Add("log filterbank", Log()),
		app.NodeByName("filterbank"),
	)

	os.Exit(m.Run())
}

func TestDFTEgyFeature(t *testing.T) {

	app.Reset()
	sp := app.NodeByName("spectrum")

	// Compare results.
	cnt := 0
	for i := 0; i < nf; i++ {
		value, e := sp.Get(i)
		if e != nil {
			t.Fatal(e)
		}
		f := value.(*narray.NArray).Data

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
	feat := app.NodeByName("filterbank")
	t.Log("fb", fb)
	// Compare results.
	cnt := 0
	for i := 0; i < nf; i++ {
		value, e := feat.Get(i)
		if e != nil {
			t.Fatal(e)
		}
		f := value.(*narray.NArray).Data

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
	feat := app.NodeByName("log filterbank")
	t.Log("logfb", logfb)

	// Compare results.
	cnt := 0
	for i := 0; i < nf; i++ {
		value, e := feat.Get(i)
		if e != nil {
			t.Fatal(e)
		}
		f := value.(*narray.NArray).Data

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

func wavSP(w []float64) dsp.Processer {
	return dsp.NewProc(1, func(idx int, in ...dsp.Processer) (dsp.Value, error) {
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

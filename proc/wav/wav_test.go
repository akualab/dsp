package wav

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/akualab/dsp"
	"github.com/akualab/dsp/proc"
	narray "github.com/akualab/narray/na64"
)

var dir = "../../data"

func TestJSONStreamer(t *testing.T) {

	ref := map[string]bool{"wav1": true, "wav2": true}
	ids := []string{}
	iter, err := NewIterator(dir, 8000, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	i := 0
	for ; ; i++ {
		w, e := iter.Next()
		if e == Done {
			break
		}
		if e != nil {
			t.Fatal(e)
		}
		ids = append(ids, w.ID)
		nf := iter.NumFrames()
		t.Log(i, "wav:", w.ID, "num_frames:", nf)
		if nf != len(w.Samples)/2 {
			t.Fatalf("num frames mismatch - expected %d, got %d", len(w.Samples)/2, nf)
		}
	}
	e := iter.Close()
	if e != nil {
		t.Fatal(e)
	}

	if i != 2 {
		t.Fatalf("expected 2 waveforms, got %d", i)
	}

	for _, v := range ids {
		_, ok := ref[v]
		if !ok {
			t.Fatalf("bad id %s", v)
		}
	}
}

func TestBounds(t *testing.T) {

	ref := map[string]bool{"wav1": true, "wav2": true}
	ids := []string{}
	iter, err := NewIterator(dir, 8000, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	i := 0
	for ; ; i++ {
		w, e := iter.Next()
		if e == Done {
			break
		}
		if e != nil {
			t.Fatal(e)
		}
		ids = append(ids, w.ID)
		nf := iter.NumFrames()
		t.Log(i, "wav:", w.ID, "num_frames:", nf)
		if nf != len(w.Samples)/2 {
			t.Fatalf("num frames mismatch - expected %d, got %d", len(w.Samples)/2, nf)
		}
		frame, err := iter.Frame(11)
		if err != nil {
			t.Fatal("expected frame 11, got error")
		}
		t.Log(frame)
		frame, err = iter.Frame(8000)
		if err == nil {
			t.Fatalf("expected error for frame 11, got %v", frame)
		}
		if err != dsp.ErrOOB {
			t.Fatalf("expected [%s] got %s", dsp.ErrOOB, err)
		}
	}
	e := iter.Close()
	if e != nil {
		t.Fatal(e)
	}

	if i != 2 {
		t.Fatalf("expected 2 waveforms, got %d", i)
	}

	for _, v := range ids {
		_, ok := ref[v]
		if !ok {
			t.Fatalf("bad id %s", v)
		}
	}
}

func ExampleNewSourceProc_spectrum() {

	app := dsp.NewApp("Example App")

	// Read a waveform.
	path := filepath.Join(dir, "wav1.json.gz")
	wavSource, err := NewSourceProc(path, Fs(8000))
	if err != nil {
		panic(err)
	}

	// Add the source processor responsible for reading and supplying waveform samples.
	wav := app.Add("wav", wavSource)

	// Use a windowing processor to segment the waveform into frames of 80 samples
	// and apply a Hamming window of size 205. The last arg is to instruct the processor
	// to center the frame in the middle of the window.
	window := app.Add("window", proc.NewWindowProc(80, 205, proc.Hamming, true))

	// Compute the FFT of the windowed frame. The FFT size is 2**8.
	spectrum := app.Add("spectrum", proc.SpectralEnergy(8))

	// Connect the processors.
	// wav -> window -> spectrum
	app.Connect(window, wav)
	app.Connect(spectrum, window)

	// Get features using this object.
	out := spectrum

	// Get the next waveform. (This processor is designed to read a list of
	// files. The Next() method loads the next waveform in the list.)
	wavSource.Next()

	// Get the spectrum for frame #10.
	v, e := out.Get(10)
	if e != nil {
		panic(e)
	}

	// Print first element of the FFT for frame #10.
	fmt.Println(v.(*narray.NArray).Data[0])
	// Output:
	// 5.123420120893221e-05
}

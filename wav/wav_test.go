package wav

import (
	"testing"

	"github.com/akualab/dsp"
)

func TestJSONStreamer(t *testing.T) {

	ref := map[string]bool{"wav1": true, "wav2": true, "short": true}
	ids := []string{}
	iter, err := NewIterator("../data", 8000, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("XXX", iter)
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
		t.Log("XXX", iter, nf, w.ID)
		t.Log(i, "wav:", w.ID, "num_frames:", nf)
		if nf != len(w.Samples)/2 {
			t.Fatalf("num frames mismatch - expected %d, got %d", len(w.Samples)/2, nf)
		}
	}
	e := iter.Close()
	if e != nil {
		t.Fatal(e)
	}

	if i != 3 {
		t.Fatalf("expected 3 waveforms, got %d", i)
	}

	for _, v := range ids {
		_, ok := ref[v]
		if !ok {
			t.Fatalf("bad id %s", v)
		}
	}
}

func TestBounds(t *testing.T) {

	ref := map[string]bool{"wav1": true, "wav2": true, "short": true}
	ids := []string{}
	iter, err := NewIterator("../data", 8000, 2, 2)
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

	if i != 3 {
		t.Fatalf("expected 3 waveforms, got %d", i)
	}

	for _, v := range ids {
		_, ok := ref[v]
		if !ok {
			t.Fatalf("bad id %s", v)
		}
	}
}

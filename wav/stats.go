// Copyright (c) 2015 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wav

import "math"

func (w *Waveform) stats() {

	var sumx, sumxsq float64
	for _, v := range w.Samples {
		sumx += v
		sumxsq += v * v
	}
	w.sumx = sumx
	w.sumxsq = sumxsq
}

// Mean returns the mean of the waveform samples as they were read from the source.
func (w *Waveform) Mean() float64 {
	return w.sumx / float64(len(w.Samples))
}

// SD returns the standard deviation of the waveform samples as they were read from the source.
func (w *Waveform) SD() float64 {
	n := float64(len(w.Samples))
	mu := w.sumx / n
	return math.Sqrt(w.sumxsq/n - mu*mu)
}

// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proc

import (
	"testing"
)

func comparef64(f1, f2, epsilon float64) bool {
	err := f2 - f1
	if err < 0 {
		err = -err
	}
	if err < epsilon {
		return true
	}
	return false
}

func compareSliceFloat(t *testing.T, expected []float64, actual []float64, message string, epsilon float64) {
	for i, _ := range expected {
		if !comparef64(expected[i], actual[i], epsilon) {
			t.Errorf("[%s]. Expected: [%f], Got: [%f]",
				message, expected[i], actual[i])
		}
	}
}

func compareFloats(t *testing.T, expected float64, actual float64, message string, epsilon float64) {
	if !comparef64(expected, actual, epsilon) {
		t.Errorf("[%s]. Expected: [%f], Got: [%f]",
			message, expected, actual)
	}
}

func compareSliceInt(t *testing.T, expected []int, actual []int, message string) {
	for i, _ := range expected {
		if expected[i] != actual[i] {
			t.Errorf("[%s]. Expected: [%d], Got: [%d]",
				message, expected[i], actual[i])
		}
	}
}

func CheckError(t *testing.T, e error) {

	if e != nil {
		t.Fatal(e)
	}
}

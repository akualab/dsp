// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsp

import (
	"bufio"
	"fmt"
	"io"
)

type ValueType int

const (
	bufferSize = 10000

	Text ValueType = iota
	Float64
	Float32
	Int32
	Int16
	Int8
)

// WriteValues prints each input vector v followed by a newline to
// writer; and in addition it emits v.  Therefore WriteValues()
// can be used like the "tee" command, which can often be useful
// for debugging.
func WriteValues(writer io.Writer, on bool) Processer {
	return NewProc(defaultBufSize, func(idx int, in ...Processer) (Value, error) {
		v, err := Processers(in).Get(idx)
		if err != nil {
			return nil, err
		}
		b := bufio.NewWriter(writer)
		if on {
			s := fmt.Sprintf("%v\n", v)
			if _, err := b.Write([]byte(s)); err != nil {
				panic(err)
			}
		}
		b.Flush()
		return v, nil
	})
}

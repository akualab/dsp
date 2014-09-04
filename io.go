// Copyright (c) 2014 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsp

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
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
func WriteValues(writer io.Writer, on bool) Processor {
	return ProcFunc(func(arg Arg) error {
		b := bufio.NewWriter(writer)
		for v := range arg.In[0] {
			if on {
				s := fmt.Sprintf("%v\n", v)
				if _, err := b.Write([]byte(s)); err != nil {
					return err
				}
			}
			SendValue(v, arg)
			//arg.Out <- v
		}
		return b.Flush()
	})
}

type ReaderConfig struct {
	// The length in samples of the output vector.
	FrameSize int
	// The overlap in samples between successive vectors.
	// Default is FrameSize (no overlap).
	StepSize int
	// The input data format. Can be text with separator or a binary representation.
	ValueType ValueType
	// Binary order. Depends on archgitecture. Default is
	ByteOrder binary.ByteOrder
}

// Returns a Reader Processor with default values: text format, no overalp, little-endian.
// Word delimiter for text format is space, any of: '\t', '\n', '\v', '\f', '\r', ' ', U+0085 (NEL), U+00A0 (NBSP)
// A frame shifts values from previous frame to the left and adds "step size" new values.
func NewReader(frameSize int) *ReaderConfig {
	return &ReaderConfig{
		FrameSize: frameSize,
		StepSize:  frameSize,
		ValueType: Text,
		ByteOrder: binary.LittleEndian,
	}
}

// Processor to read values from an io.Reader interface.
func Reader(reader io.Reader, config *ReaderConfig) Processor {
	return ProcFunc(func(arg Arg) error {

		if config.StepSize > config.FrameSize {
			return fmt.Errorf("bad step size [%d] is greater than frame size [%d]", config.StepSize, config.FrameSize)
		}
		fs := config.FrameSize
		frame := make(Value, fs, fs)
		ovs := fs - config.StepSize // zero if no overlap
		overlap := make(Value, ovs, ovs)
		b := bufio.NewReaderSize(reader, bufferSize)
		switch config.ValueType {

		case Text:
			scanner := bufio.NewScanner(b)
			scanner.Split(bufio.ScanWords)

			i := ovs
			for scanner.Scan() {
				v, e := strconv.ParseFloat(scanner.Text(), 64)
				if e != nil {
					return e
				}
				frame[i] = v
				i++
				if i >= fs {
					// Pad with overlap data
					copy(frame[0:ovs], overlap)
					// Save overlap dat for next frame.
					copy(overlap, frame[config.StepSize:])
					// Frame ready, send out.
					SendValue(frame, arg)
					// Prepare for next frame.
					i = ovs
					frame = make(Value, fs, fs)
				}
			}
			if err := scanner.Err(); err != nil {
				fmt.Fprintln(os.Stderr, "reading input:", err)
			}

		case Float64:
		case Float32:
		case Int32:
		case Int16:
		case Int8:
		default:
			return fmt.Errorf("unknown value type: %d", config.ValueType)
		}

		return nil
	})
}

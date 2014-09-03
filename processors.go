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

func Split(writer io.Writer, on bool) Processor {
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
		}
		return b.Flush()
	})
}

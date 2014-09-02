package stream

import (
	"bufio"
	"fmt"
	"io"
)

// WriteValues prints each input vector v followed by a newline to
// writer; and in addition it emits v.  Therefore WriteValues()
// can be used like the "tee" command, which can often be useful
// for debugging.
func WriteValues(writer io.Writer) Processor {
	return ProcFunc(func(arg Arg) error {
		b := bufio.NewWriter(writer)
		for v := range arg.In[0] {
			s := fmt.Sprintf("%v\n", v)
			if _, err := b.Write([]byte(s)); err != nil {
				return err
			}
			arg.Out <- v
		}
		return b.Flush()
	})
}

package stream

import "math/rand"

// Random emits a finite sequence of random vectors.
// size is the size of the vector.
// length is the length of the sequence.
func Random(size, length int) Processor {
	return ProcFunc(func(arg Arg) error {
		r := rand.New(rand.NewSource(99))
		for i := 0; i < length; i++ {
			v := make(Value, size, size)
			for j := 0; j < size; j++ {
				v[j] = r.Float64()
			}
			arg.Out <- v
		}
		return nil
	})
}

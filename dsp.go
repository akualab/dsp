package dsp

import (
	"bytes"
	"errors"
	"fmt"
)

// Called a Proc that has no ProcFunc set.
var ErrNoFunc = errors.New("no ProcFunc set")

// Returned when frame index is out of bounds. Can be used as a termination flag.
var ErrOOB = errors.New("frame index out of bounds")

// Value is the type used to exchange values between processors.
type Value []float64

// The Processer interface must be implemented by all processors.
type Processer interface {
	Get(uint32) (Value, error)
	SetInputs(...Processer)
	Reset()
}

// ProcFunc is the type used to implement processing functions.
type ProcFunc func(uint32, ...Processer) (Value, error)

// Proc can be embedded in objects that implement the Processer interface.
type Proc struct {
	f      ProcFunc
	inputs []Processer
	cache  *cache
}

// NewProc creates a new Proc.
func NewProc(bufSize int, f ProcFunc) *Proc {
	return &Proc{
		f:     f,
		cache: newCache(bufSize),
	}
}

// SetInputs sets the inputs for a processor.
func (bp *Proc) SetInputs(inputs ...Processer) {
	bp.inputs = inputs
}

// Reset - override this method to reset the processor state.
func (bp *Proc) Reset() {
	bp.cache.clear()
}

// Get - returns value for index.
func (bp *Proc) Get(idx uint32) (Value, error) {
	val, ok := bp.cache.get(idx)
	if ok {
		return val, nil
	}
	if bp.f != nil {
		v, e := bp.f(idx, bp.inputs...)
		if e != nil {
			return nil, e
		}
		bp.cache.set(idx, v)
		return v, nil
	}
	return nil, ErrNoFunc
}

// SetCache sets the value in the cache.
func (bp *Proc) SetCache(idx uint32, val Value) {
	bp.cache.set(idx, val)
}

// GetCache gets value from cache.
func (bp *Proc) GetCache(idx uint32) (Value, bool) {
	val, ok := bp.cache.get(idx)
	return val, ok
}

// ClearCache clears the cache.
func (bp *Proc) ClearCache() {
	bp.cache.clear()
}

// Inputs returns the input processors.
func (bp *Proc) Inputs() []Processer {
	return bp.inputs
}

// Input returns one of the processor inputs.
func (bp *Proc) Input(n int) Processer {
	return bp.inputs[n]
}

// App defines a DSP application.
type App struct {
	// App name.
	Name string
	// Default buffer size for connecting channels.
	procs  map[string]Processer
	inputs map[string][]string
}

// NewApp returns a new app.
func NewApp(name string) *App {
	return &App{
		Name:   name,
		procs:  make(map[string]Processer),
		inputs: make(map[string][]string),
	}
}

func (app *App) mustGet(name string) Processer {
	proc, ok := app.procs[name]
	if !ok {
		panic(fmt.Errorf("no processor named [%s] in builder graph", name))
	}
	return proc
}

// Tap provides access to values in the processor graph.
type Tap struct {
	proc Processer
}

// NewTap returns a tap for the output of the named processor.
func (app *App) NewTap(name string) Tap {
	return Tap{
		proc: app.mustGet(name),
	}
}

// Value returns the value for frame index.
func (t Tap) Get(idx uint32) (Value, error) {
	return t.proc.Get(idx)
}

// Add adds a processor with a name.
func (app *App) Add(name string, p Processer) string {
	app.procs[name] = p
	return name
}

// Connect connects processor inputs. Example:
//    app.Connect("y", "x1", "x2")
// the output values of processors with name "x1" and "x2" are
// inputs to processor names "y".
func (app *App) Connect(to string, from ...string) {

	inputs := []Processer{}
	toProc := app.mustGet(to)
	for _, in := range from {
		input := app.mustGet(in)
		inputs = append(inputs, input)
	}
	toProc.SetInputs(inputs...)
	app.inputs[to] = from
}

// Reset resets all processors in preparation fro a new stream.
func (app *App) Reset() {
	for _, p := range app.procs {
		p.Reset()
	}
}

func (app *App) String() string {

	var buf bytes.Buffer
	for name, _ := range app.procs {
		buf.WriteString(fmt.Sprintf("proc: %s, inputs: | ", name))
		for _, in := range app.inputs[name] {
			buf.WriteString(fmt.Sprintf("%s | ", in))
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

// Copy creates a copy of the value.
// Input values should be treated as read-only because
// they may be shared with other processors.
func (v Value) Copy() Value {
	vcopy := make(Value, len(v), len(v))
	copy(vcopy, v)
	return vcopy
}

package signals

import (
	"encoding/gob"
	"io"
)

func init() {
	gob.Register(Modulated{})
	gob.Register(Composite{})
	gob.Register(Stack{})
}

// Modulated is a PeriodicLimitedSignal, generated by multiplying together Signal(s).(Signal's can be PeriodicLimitedSignal's, so this can be hierarchical.)
// Multiplication scales so that, Maxy*Maxy=Maxy (so Maxy is unity).
// Modulated's MaxX() comes from the smallest contstituent MaxX(), (0 if none of the contained Signals are LimitedSignals.)
// Modulated's Period() comes from its first member.
// As with 'AND' logic, all sources have to be Maxy (at a particular x) for Modulated to be Maxy, whereas, ANY Signal at zero will generate a Modulated of zero.
type Modulated []Signal

func (c Modulated) property(t x) (total y) {
	total = unitY
	for _, s := range c {
		l := s.property(t)
		switch l {
		case 0:
			total = 0
			break
		case unitY:
			continue
		default:
			//total = (total / Halfy) * (l / Halfy)*2
			total = (total >> halfyBits) * (l >> halfyBits) * 2
		}
	}
	return
}

func (c Modulated) Period() (period x) {
	// TODO needs to be longest period and all constituents but only when the shorter are multiples of it.
	if len(c) > 0 {
		if s, ok := c[0].(PeriodicSignal); ok {
			return s.Period()
		}
	}
	return
}

// the smallest Max X of the constituents.
func (c Modulated) MaxX() (min x) {
	min = -1
	for _, s := range c {
		if sls, ok := s.(LimitedSignal); ok {
			if newmin := sls.MaxX(); newmin >= 0 && (min == -1 || newmin < min) {
				min = newmin
			}
		}
	}
	return
}

// write Gob encoding
func (c Modulated) Save(p io.Writer) error {
	return gob.NewEncoder(p).Encode(&c)
}

// read Gob encoding
func (c *Modulated) Load(p io.Reader) error {
	return gob.NewDecoder(p).Decode(c)
}

// returns a PeriodicLimitedSignal (type Modulated) based on a sine wave,
// with peak y set to Maxy adjusted by dB,
// so dB should always be negative.
func NewTone(period x, dB float64) Modulated {
	return Modulated{Sine{period}, NewConstant(dB)}
}

// helper to enable generation from another slice.
// will in general need to use a slice interface promoter function.
func NewModulated(c ...Signal) Modulated {
	return Modulated(c)
}

// Composite is a PeriodicLimitedSignal, generated by adding together Signal(s). (PeriodicLimitedSignal's are Signal's so this can be hierarchical.)
// Composite's MaxX() comes from the largest contstituent MaxX(), (0 if none of the contained Signals are LimitedSignals.)
// Composite's Period() comes from its first member.
// As with 'OR' logic, all sources have to be zero (at a particular x) for Composite to be zero.
type Composite []Signal

func (c Composite) property(t x) (total y) {
	for _, s := range c {
		total += s.property(t)
	}
	return
}

func (c Composite) Period() (period x) {
	// TODO could helpfully be the longest period and all constituents but only when the shorter are multiples of it.
	if len(c) > 0 {
		if s, ok := c[0].(PeriodicSignal); ok {
			return s.Period()
		}
	}
	return
}

// the largest Max X of the constituents.
func (c Composite) MaxX() (max x) {
	max = -1
	for _, s := range c {
		if sls, ok := s.(LimitedSignal); ok {
			if newmax := sls.MaxX(); newmax > max {
				max = newmax
			}
		}
	}
	return
}

// write Gob encoding
func (c Composite) Save(p io.Writer) error {
	return gob.NewEncoder(p).Encode(&c)
}

// read Gob encoding
func (c *Composite) Load(p io.Reader) error {
	return gob.NewDecoder(p).Decode(c)
}

// helper to enable generation from another slice.
// will in general need to use a slice interface promoter function.
func NewComposite(c ...Signal) Composite {
	return Composite(c)
}

// Stack is a PeriodicLimitedSignal, generated by adding together Signal(s).(Signal's can be PeriodicLimitedSignal's, so this can be hierarchical.)
// Stack's MaxX() comes from the largest contstituent MaxX(), (0 if no contained Signals are LimitedSignals.)
// Stack's Period() comes from its first member.
// Overwise like Composite, Stack scales down by len(Stack), making it impossible to overrun maxy.
// As with 'OR' logic, all sources have to be zero (at a particular x) for Stack to be zero.
type Stack []Signal

func (c Stack) property(t x) (total y) {
	for _, s := range c {
		total += s.property(t) / y(len(c))
	}
	return
}

func (c Stack) Period() (period x) {
	// TODO needs to be longest period and all constituents but only when the shorter are multiples of it.
	if len(c) > 0 {
		if s, ok := c[0].(PeriodicSignal); ok {
			return s.Period()
		}
	}
	return
}

// the largest Max X of the constituents.
func (c Stack) MaxX() (max x) {
	max = -1
	for _, s := range c {
		if sls, ok := s.(LimitedSignal); ok {
			if newmax := sls.MaxX(); newmax > max {
				max = newmax
			}
		}
	}
	return
}

// write Gob encoding
func (c Stack) Save(p io.Writer) error {
	return gob.NewEncoder(p).Encode(&c)
}

// read Gob encoding
func (c *Stack) Load(p io.Reader) error {
	return gob.NewDecoder(p).Decode(c)
}

// helper to enable generation from another slice.
// will in general need to use a slice interface promoter function.
func NewStack(c ...Signal) Stack {
	return Stack(c)
}

// Converters to promote slices of interfaces, needed when using variadic parameters called using a slice since go doesn't automatically promote a narrow interface inside the slice to be able to use a broader interface.
// for example: without these you couldn't use a slice of LimitedSignal's in a variadic call to a func requiring Signal's. (when you can use separate LimitedSignal's in the same call.)

// converts to []Signal
func PromoteToSignals(s interface{}) (out []Signal) {
	switch st := s.(type) {
	case []LimitedSignal:
		out = make([]Signal, len(st))
		for i := range out {
			out[i] = st[i].(Signal)
		}
	case []PeriodicLimitedSignal:
		out = make([]Signal, len(st))
		for i := range out {
			out[i] = st[i].(Signal)
		}
	case []PeriodicSignal:
		out = make([]Signal, len(st))
		for i := range out {
			out[i] = st[i].(Signal)
		}
	case []PCMSignal:
		out = make([]Signal, len(st))
		for i := range out {
			out[i] = st[i].(Signal)
		}
	}
	return
}

// converts to []LimitedSignal
func PromoteToLimitedSignals(s interface{}) (out []LimitedSignal) {
	switch st := s.(type) {
	case []PeriodicLimitedSignal:
		out = make([]LimitedSignal, len(st))
		for i := range out {
			out[i] = st[i].(LimitedSignal)
		}
	case []PCMSignal:
		out = make([]LimitedSignal, len(st))
		for i := range out {
			out[i] = st[i].(LimitedSignal)
		}
	}
	return
}

// converts to []PeriodicSignal
func PromoteToPeriodicSignals(s interface{}) (out []PeriodicSignal) {
	switch st := s.(type) {
	case []PeriodicLimitedSignal:
		out = make([]PeriodicSignal, len(st))
		for i := range out {
			out[i] = st[i].(PeriodicSignal)
		}
	case []PCMSignal:
		out = make([]PeriodicSignal, len(st))
		for i := range out {
			out[i] = st[i].(PeriodicSignal)
		}
	}
	return
}

// converts to []PeriodicLimitedSignal
func PromoteToPeriodicLimitedSignals(s interface{}) (out []PeriodicLimitedSignal) {
	switch st := s.(type) {
	case []PCMSignal:
		out = make([]PeriodicLimitedSignal, len(st))
		for i := range out {
			out[i] = st[i].(PeriodicLimitedSignal)
		}
	}
	return
}

package signals

import (
	"encoding/gob"
	"io"
)

func init() {
	gob.Register(Modulated{})
	gob.Register(Compose{})
	gob.Register(Stack{})
}

// Modulated is a PeriodicLimitedFunction, generated by multiplying together Function(s).
// multiplication scales so that, Maxy*Maxy=Maxy (so Maxy is unity).
// Modulated's MaxX() comes from the smallest contstituent MaxX(), (0 if no contained Functions are LimitedFunctions.)
// Modulated's Period() comes from its first member.
// as with 'AND' logic, all sources have to be Maxy (at a particular momemt) for Modulated to be Maxy, whereas, ANY function at zero will generate a Modulated of zero.
type Modulated []Function

func (c Modulated) call(t x) (total y) {
	total = maxY
	for _, s := range c {
		l := s.call(t)
		switch l {
		case 0:
			total = 0
			break
		case maxY:
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
		if s, ok := c[0].(PeriodicFunction); ok {
			return s.Period()
		}
	}
	return
}

// the smallest Max X of the constituents.
func (c Modulated) MaxX() (min x) {
	min = -1
	for _, s := range c {
		if sls, ok := s.(LimitedFunction); ok {
			if newmin := sls.MaxX(); newmin >= 0 && (min == -1 || newmin < min) {
				min = newmin
			}
		}
	}
	return
}

func (c Modulated) Save(p io.Writer) error {
	return gob.NewEncoder(p).Encode(&c)
}

func (c *Modulated) Load(p io.Reader) error {
	return gob.NewDecoder(p).Decode(c)
}

// helper to enable generation from another slice.
// will generally need to use an slice interface promoter utility.
func NewMultiplex(c ...Function) Modulated {
	return Modulated(c)
}

// Compose is a PeriodicLimitedFunction, generated by adding together Function(s).
// Compose's MaxX() comes from the largest contstituent MaxX(), (0 if no contained Functions are LimitedFunctions.)
// Compose's Period() comes from its first member.
// as with 'OR' logic, all sources have to be zero (at a particular momemt) for Compose to be zero.
type Compose []Function

func (c Compose) call(t x) (total y) {
	for _, s := range c {
		total += s.call(t)
	}
	return
}

func (c Compose) Period() (period x) {
	// TODO needs to be longest period and all constituents but only when the shorter are multiples of it.
	if len(c) > 0 {
		if s, ok := c[0].(PeriodicFunction); ok {
			return s.Period()
		}
	}
	return
}

// the largest Max X of the constituents.
func (c Compose) MaxX() (max x) {
	max = -1
	for _, s := range c {
		if sls, ok := s.(LimitedFunction); ok {
			if newmax := sls.MaxX(); newmax > max {
				max = newmax
			}
		}
	}
	return
}

func (c Compose) Save(p io.Writer) error {
	return gob.NewEncoder(p).Encode(&c)
}

func (c *Compose) Load(p io.Reader) error {
	return gob.NewDecoder(p).Decode(c)
}

// helper to enable generation from another slice.
// will generally need to use an slice interface promoter utility.
func NewCompositor(c ...Function) Compose {
	return Compose(c)
}

// Stack is a PeriodicLimitedFunction, generated by adding together Function(s).
// Stack's MaxX() comes from the largest contstituent MaxX(), (0 if no contained Functions are LimitedFunctions.)
// Stack's Period() comes from its first member.
// Unlike Combine, Stack scales down by len(Stack), making it impossible to overrun maxy.
// as with 'OR' logic, all sources have to be zero (at a particular momemt) for Stack to be zero.
type Stack []Function

func (c Stack) call(t x) (total y) {
	for _, s := range c {
		total += s.call(t) / y(len(c))
	}
	return
}

func (c Stack) Period() (period x) {
	// TODO needs to be longest period and all constituents but only when the shorter are multiples of it.
	if len(c) > 0 {
		if s, ok := c[0].(PeriodicFunction); ok {
			return s.Period()
		}
	}
	return
}

// the largest Max X of the constituents.
func (c Stack) MaxX() (max x) {
	max = -1
	for _, s := range c {
		if sls, ok := s.(LimitedFunction); ok {
			if newmax := sls.MaxX(); newmax > max {
				max = newmax
			}
		}
	}
	return
}

func (c Stack) Save(p io.Writer) error {
	return gob.NewEncoder(p).Encode(&c)
}

func (c *Stack) Load(p io.Reader) error {
	return gob.NewDecoder(p).Decode(c)
}

// helper to enable generation from another slice.
// will generally need to use an slice interface promoter utility.
func NewStack(c ...Function) Stack {
	return Stack(c)
}

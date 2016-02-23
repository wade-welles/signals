package signals

import (
	"encoding/gob"
	"io"
)

func init() {
	gob.Register(Multiplex{})
	gob.Register(Sum{})
	gob.Register(Stack{})
}

// returns a periodical (type multiplex) based on a sine wave,
// with peak y set to Maxy adjusted by dB,
// dB should always be negative to remain in y limits.
func NewTone(period x, dB float64) Multiplex {
	return Multiplex{Sine{period}, NewConstant(dB)}
}

// Multiplex is a Function generated by multiplying together Function(s).
// multiplication scales so that, Maxy*Maxy=Maxy (so Maxy is unity).
// like logic AND; all its functions (at a particular momemt) need to be Maxy to produce a Multi of Maxy, whereas, ANY function at zero will generate a Multi of zero.
// Multiplex is also a Periodical, taking its period, if any, from its first member.
type Multiplex []Function

func (c Multiplex) Call(t x) (total y) {
	total = Maxy
	for _, s := range c {
		l := s.Call(t)
		switch l {
		case 0:
			total = 0
			break
		case Maxy:
			continue
		default:
			//total = (total / Halfy) * (l / Halfy)*2
			total = (total >> HalfyBits) * (l >> HalfyBits) * 2
		}
	}
	return
}

func (c Multiplex) Period() (period x) {
	// TODO needs to be longest period and all constituents but only when the shorter are multiples of it.
	if len(c) > 0 {
		if s, ok := c[0].(PeriodicFunction); ok {
			return s.Period()
		}
	}
	return
}

// the smallest Max X of the constituents.
func (c Multiplex) MaxX() (min x) {
	min = -1
	for _, s := range c {
		if sls, ok := s.(limiter); ok {
			if newmin := sls.MaxX(); newmin >= 0 && (min == -1 || newmin < min) {
				min = newmin
			}
		}
	}
	return
}

/*
// this doesn't work because Note is still a limiter (multiplex) an has durarion zero
// or TPIAW  multiplex with no limited's, results in a limited of zero Dx. not good
func (c Multiplex) Duration() (min x) {
	var found bool
	for _, s := range c {
		if sls, ok := s.(limiter); ok {
			if !found{
				min=sls.Duration()
				found=true
			}else{
				if newmin := sls.Duration(); newmin < min {
					min = newmin
				}
			}
		}
	}
	if found {return}
	return 0
}
*/
func (c Multiplex) Save(p io.Writer) error {
	return gob.NewEncoder(p).Encode(&c)
}

func (c *Multiplex) Load(p io.Reader) error {
	return gob.NewDecoder(p).Decode(c)
}

// helper: needed becasue can't use type literal with array source.
func NewMultiplex(c ...Function) Multiplex {
	return Multiplex(c)
}


// Sum is a Function generated by adding together Function(s).
// also a Periodical, taking its period, if any, from its first member.
// like 'OR' logic, all sources have to be zero (at a particular momemt) for Sum to be zero.
type Sum []Function

func (c Sum) Call(t x) (total y) {
	for _, s := range c {
		total += s.Call(t)
	}
	return
}

func (c Sum) Period() (period x) {
	// TODO needs to be longest period and all constituents but only when the shorter are multiples of it.
	if len(c) > 0 {
		if s, ok := c[0].(PeriodicFunction); ok { 
			return s.Period()
		}
	}
	return
}

// the largest Max X of the constituents.
func (c Sum) MaxX() (max x) {
	max = -1
	for _, s := range c {
		if sls, ok := s.(limiter); ok {
			if newmax := sls.MaxX(); newmax > max {
				max = newmax
			}
		}
	}
	return
}

func (c Sum) Save(p io.Writer) error {
	return gob.NewEncoder(p).Encode(&c)
}

func (c *Sum) Load(p io.Reader) error {
	return gob.NewDecoder(p).Decode(c)
}

// helper: needed becasue can't use type literal with array source.
func NewSum(c ...Function) Sum {
	return Sum(c)
}

// Stack is a Function generated by adding together Function(s).
// source Functions are scaled down by Stacks count, making it impossible to overrun maxy.
// also a Periodical, taking its period, if any, from its first member.
// like 'OR' logic, all sources have to be zero (at a particular momemt) for Stack to be zero.
type Stack []Function

func (c Stack) Call(t x) (total y) {
	for _, s := range c {
		total += s.Call(t) / y(len(c))
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
		if sls, ok := s.(limiter); ok {
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

// helper: needed becasue can't use type literal with array source.
func NewStack(c ...Function) Stack {
	return Stack(c)
}

package signals

import (
	"math"
)

type Constant struct {
	Setting Level 
}

func (s Constant) Level(t Interval) Level {
	return s.Setting
}


type Sine struct {
	Period Interval
}

func (s Sine) Level(t Interval) Level {
	return Level(math.Sin(float64(t)/float64(s.Period)*2*math.Pi) * MaxLevelfloat64)
}

type Pulse struct {
	Width Interval
}

func (s Pulse) Level(t Interval) Level {
	if t > s.Width {
		return 0
	} else {
		return MaxLevel
	}
}

type Square struct {
	Period Interval
}

func (s Square) Level(t Interval) Level {
	if t%s.Period >= s.Period/2 {
		return -MaxLevel
	} else {
		return MaxLevel
	}
}

type RampUp struct {
	Period Interval
}

func (s RampUp) Level(t Interval) Level {
	if t < 0 {
		return 0
	} else if t > s.Period {
		return MaxLevel
	} else {
		return Level(Interval(MaxLevel) / s.Period * t)
	}
}

type RampDown struct {
	Period Interval
}

func (s RampDown) Level(t Interval) Level {
	if t < 0 {
		return MaxLevel
	} else if t > s.Period {
		return 0
	} else {
		return Level(Interval(MaxLevel) / s.Period * (s.Period - t))
	}
}

type Heavyside struct {
}

func (s Heavyside) Level(t Interval) Level {
	if t < 0 {
		return 0
	}
	return MaxLevel
}

type Sigmoid struct {
	Steepness Interval
}

func (s Sigmoid) Level(t Interval) Level {
	return Level(float64(MaxLevel) / (1 + math.Exp(-float64(t)/float64(s.Steepness))))
}


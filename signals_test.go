package signals

import (
	"fmt"
	"math/big"
)

func ExampleSquare() {
	s := Square{UnitX}
	for t := x(0); t < 2*UnitX; t += UnitX / 10 {
		fmt.Print(s.Call(t))
	}
	fmt.Println()
	// Output:   100.00%   100.00%   100.00%   100.00%   100.00%  -100.00%  -100.00%  -100.00%  -100.00%  -100.00%   100.00%   100.00%   100.00%   100.00%   100.00%  -100.00%  -100.00%  -100.00%  -100.00%  -100.00%
}

func ExamplePulse() {
	s := Pulse{UnitX}
	for t := x(0); t < 2*UnitX; t += UnitX / 10 {
		fmt.Print(s.Call(t))
	}
	fmt.Println()
	// Output:   100.00%   100.00%   100.00%   100.00%   100.00%   100.00%   100.00%   100.00%   100.00%   100.00%   100.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%
}
func ExampleRamp() {
	s := RampUp{UnitX}
	for t := x(0); t < 2*UnitX; t += UnitX / 10 {
		fmt.Print(s.Call(t))
	}
	fmt.Println()
	s2 := RampDown{UnitX}
	for t := x(0); t < 2*UnitX; t += UnitX / 10 {
		fmt.Print(s2.Call(t))
	}
	fmt.Println()
	// Output:   0.00%    10.00%    20.00%    30.00%    40.00%    50.00%    60.00%    70.00%    80.00%    90.00%   100.00%   100.00%   100.00%   100.00%   100.00%   100.00%   100.00%   100.00%   100.00%   100.00%
	//    100.00%    90.00%    80.00%    70.00%    60.00%    50.00%    40.00%    30.00%    20.00%    10.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%
}

func ExampleSine() {
	s := Sine{UnitX}
	for t := x(0); t < 2*UnitX; t += UnitX / 16 {
		fmt.Print(s.Call(t))
	}
	fmt.Println()
	// Output:     0.00%    38.27%    70.71%    92.39%   100.00%    92.39%    70.71%    38.27%     0.00%   -38.27%   -70.71%   -92.39%  -100.00%   -92.39%   -70.71%   -38.27%    -0.00%    38.27%    70.71%    92.39%   100.00%    92.39%    70.71%    38.27%     0.00%   -38.27%   -70.71%   -92.39%  -100.00%   -92.39%   -70.71%   -38.27%
}

func ExampleSigmoid() {
	s := Sigmoid{UnitX}
	for t := x(-2 * UnitX); t < 2*UnitX; t += UnitX / 10 {
		fmt.Print(s.Call(t))
	}
	fmt.Println()
	// Output:   11.92%    13.01%    14.19%    15.45%    16.80%    18.24%    19.78%    21.42%    23.15%    24.97%    26.89%    28.91%    31.00%    33.18%    35.43%    37.75%    40.13%    42.56%    45.02%    47.50%    50.00%    52.50%    54.98%    57.44%    59.87%    62.25%    64.57%    66.82%    69.00%    71.09%    73.11%    75.03%    76.85%    78.58%    80.22%    81.76%    83.20%    84.55%    85.81%    86.99%
}

func ExampleADSREnvelope() {
	s := NewADSREnvelope(UnitX, UnitX, UnitX, Maxy/2, UnitX)
	for t := x(0); t < 5*UnitX; t += UnitX / 10 {
		fmt.Print(s.Call(t))
	}
	fmt.Println()
	// Output:    0.00%    10.00%    20.00%    30.00%    40.00%    50.00%    60.00%    70.00%    80.00%    90.00%   100.00%    95.00%    90.00%    85.00%    80.00%    75.00%    70.00%    65.00%    60.00%    55.00%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    45.00%    40.00%    35.00%    30.00%    25.00%    20.00%    15.00%    10.00%     5.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%
}

func ExampleReflect() {
	s := Reflected{NewADSREnvelope(UnitX, UnitX, UnitX, Maxy/2, UnitX)}
	for t := x(0); t < 5*UnitX; t += UnitX / 10 {
		fmt.Print(s.Call(t))
	}
	fmt.Println()
	// Output:   100.00%    90.00%    80.00%    70.00%    60.00%    50.00%    40.00%    30.00%    20.00%    10.00%     0.00%     5.00%    10.00%    15.00%    20.00%    25.00%    30.00%    35.00%    40.00%    45.00%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    55.00%    60.00%    65.00%    70.00%    75.00%    80.00%    85.00%    90.00%    95.00%   100.00%   100.00%   100.00%   100.00%   100.00%   100.00%   100.00%   100.00%   100.00%   100.00%
}

func ExampleMultiplex() {
	s := Multiplex{Sine{UnitX * 5}, Sine{UnitX * 10}}
	for t := x(0); t < 5*UnitX; t += UnitX / 10 {
		fmt.Print(s.Call(t))
	}
	fmt.Println()
	// Output:     0.00%     0.79%     3.12%     6.90%    11.98%    18.16%    25.20%    32.81%    40.68%    48.48%    55.90%    62.61%    68.32%    72.75%    75.69%    76.94%    76.40%    73.99%    69.72%    63.65%    55.90%    46.66%    36.16%    24.67%    12.51%     0.00%   -12.51%   -24.67%   -36.16%   -46.66%   -55.90%   -63.65%   -69.72%   -73.99%   -76.40%   -76.94%   -75.69%   -72.75%   -68.32%   -62.61%   -55.90%   -48.48%   -40.68%   -32.81%   -25.20%   -18.16%   -11.98%    -6.90%    -3.12%    -0.79%
}

func ExampleStack() {
	s := Stack{Sine{UnitX * 5}, Sine{UnitX * 10}}
	for t := x(0); t < 5*UnitX; t += UnitX / 10 {
		fmt.Print(s.Call(t))
	}
	fmt.Println() // Output:     0.00%     9.41%    18.70%    27.78%    36.52%    44.84%    52.63%    59.81%    66.30%    72.03%    76.94%    80.99%    84.13%    86.35%    87.64%    88.00%    87.46%    86.03%    83.77%    80.72%    76.94%    72.52%    67.52%    62.04%    56.17%    50.00%    43.63%    37.17%    30.71%    24.34%    18.16%    12.26%     6.72%     1.60%    -3.02%    -7.10%   -10.59%   -13.45%   -15.67%   -17.24%   -18.16%   -18.45%   -18.13%   -17.24%   -15.82%   -13.94%   -11.65%    -9.04%    -6.17%    -3.13%
}

func ExampleTrigger() {
	s := Triggered{NewADSREnvelope(UnitX, UnitX, UnitX, Maxy/2, UnitX), Maxy / 3 * 2, true, UnitX / 100, UnitX * 10, 0, nil, 0, false}
	for t := x(0); t < 5*UnitX; t += UnitX / 10 {
		fmt.Print(s.Call(t))
	}
	fmt.Println()
	fmt.Println(s.Shift)
	//s.Trigger = Maxy / 3
	s.Rising = false
	for t := x(0); t < 5*UnitX; t += UnitX / 10 {
		fmt.Print(s.Call(t))
	}
	fmt.Println()
	fmt.Println(s.Shift)
	fmt.Println()
	// Output:    67.00%    77.00%    87.00%    97.00%    96.50%    91.50%    86.50%    81.50%    76.50%    71.50%    66.50%    61.50%    56.50%    51.50%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    46.50%    41.50%    36.50%    31.50%    26.50%    21.50%    16.50%    11.50%     6.50%     1.50%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%
	//      0.67
	//     66.50%    61.50%    56.50%    51.50%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    50.00%    46.50%    41.50%    36.50%    31.50%    26.50%    21.50%    16.50%    11.50%     6.50%     1.50%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%
	//      1.67
}
func ExampleNoise() {
	s := NewNoise()
	for t := x(0); t < 40*UnitX; t += UnitX {
		fmt.Print(s.Call(t))
	}
	fmt.Println()
	// Output:    23.94%   -52.49%     8.21%    -9.87%   -74.46%   -68.54%   -31.13%   -28.89%    11.03%    43.01%   -71.97%   -35.88%   -58.86%    47.80%    21.68%   -34.58%   -66.41%    10.38%     4.28%   -14.14%   -17.82%   -31.24%    22.84%   -21.90%    17.72%    23.27%    38.15%    65.67%   -72.58%   -66.54%   -33.93%     4.60%   -42.08%   -36.43%   -48.60%   -10.65%   -17.75%    25.50%    23.76%   -87.69%
}
func ExampleBitPulses() {
	i := new(big.Int)
	_, err := fmt.Sscanf("01110111011101110111011101110111", "%b", i)
	if err != nil {
		panic(i)
	}
	s := PulsePattern{*i, UnitX}
	for t := x(0); t < 40*UnitX; t += UnitX {
		fmt.Print(s.Call(t))
	}
	fmt.Println()
	// Output:   100.00%   100.00%   100.00%     0.00%   100.00%   100.00%   100.00%     0.00%   100.00%   100.00%   100.00%     0.00%   100.00%   100.00%   100.00%     0.00%   100.00%   100.00%   100.00%     0.00%   100.00%   100.00%   100.00%     0.00%   100.00%   100.00%   100.00%     0.00%   100.00%   100.00%   100.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%     0.00%
}

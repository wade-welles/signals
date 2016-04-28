// command for piping from gob encodings of functions to PCM.
// example usage (to play a tone):-
// ./player\[SYSV64\].elf < gobs/1kSine.gob | aplay
// or
// cat gobs/1kSine.gob | ./player\[SYSV64\].elf | aplay
//  (1kSine.gob is a procedural 1k cycles sine wave.)
// to specifiy duration:
// ./player\[SYSV64\].elf -length=2 < 1kSine.gob | aplay
// to specifiy sample rate:
// ./player\[SYSV64\].elf -rate=16000 < 1kSine.gob | aplay
// (output s not a higher frequency, since player passes wave format and so includes rate.)
// to specifiy sample precision:
// ./player\[SYSV64\].elf -bytes=1 < 1kSine.gob | aplay
// (bytes can be one of: 1,2,3,4.)
package main

import (
	"bufio"
	"flag"
	"os"
)

import signals "github.com/splace/signals"


func main() {
	help := flag.Bool("help", false, "display help/usage.")
	var sampleRate uint
	flag.UintVar(&sampleRate, "rate", 8000, "`samples` per unit.")
	var samplePrecision uint
	flag.UintVar(&samplePrecision, "bytes", 2, "`bytes` per sample.")
	var length float64
	flag.Float64Var(&length, "length", 1, "length in `units`")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}
	m1 := signals.Modulated{}
	rr := bufio.NewReader(os.Stdin)
	if err := m1.Load(rr); err != nil {
		panic("unable to load."+err.Error())
	}
	signals.Encode(os.Stdout,m1,signals.X(length),uint32(sampleRate),uint8(samplePrecision))
	os.Stdout.Close()
}

package signals

import (
	"io"
	"net/http"
	"errors"
)

const bufferSize = 2880   // multiple of ALL possible PCM types. (so always whole number of samples.)

type Wav struct {
	PeriodicLimitedSignal
	reader io.Reader
	buf []byte
	shift *x
}

func NewWav(URL string) (*Wav, error) {
	r, sampleBytes, sampleRate, err := PCMReader(URL)
	if err != nil {
		return nil, err
	}
	//b:=bufio.NewReaderSize(reader,bufferSize)
	b:=make([]byte,bufferSize)
	_,err=r.Read(b)
	var s x
	switch sampleBytes {
	case 1:
		return &Wav{NewPCM8bit(sampleRate, b), r,b,&s},nil
	case 2:
		return &Wav{NewPCM16bit(sampleRate,  b), r,b,&s},nil
	case 3:
		return &Wav{NewPCM24bit(sampleRate,  b), r,b,&s},nil
	case 4:
		return &Wav{NewPCM32bit(sampleRate,  b), r,b,&s},nil
	case 6:
		return &Wav{NewPCM48bit(sampleRate,  b), r,b,&s},nil
	}
	return nil,ErrWavParse{"Source bit rate not supported."}
}

func (s Wav) property(offset x) y {
	if offset>*s.shift+s.MaxX(){
		n,err:=s.reader.Read(s.buf)  
		if n<len(s.buf) || err != nil {
			//if err == EOF s.r.Close()
			panic(err)
		}
		*s.shift=*s.shift+s.MaxX()
	}
	return s.PeriodicLimitedSignal.property(offset-*s.shift)
}

func PCMReader(source string) (io.Reader, uint16, uint32, error) {
	resp, err := http.Get(source)
	if err != nil {
		return nil, 0, 0, err
	}
	//fmt.Println(resp.Header)
	if resp.Header["Content-Type"][0] == "sound/wav" {
		_, format, err := readHeader(resp.Body)
	if err != nil {
		return nil, 0, 0, err
	}
		return resp.Body, format.SampleBytes, format.SampleRate, nil
	}
	if resp.Header["Content-Type"][0] == "audio/l16;rate=8000" {
		return resp.Body, 2, 8000, nil
	}
	return nil, 0, 0, errors.New("Source in unrecognized format.")
}





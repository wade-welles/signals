package signals

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// encode a signal as PCM data in a Riff wave container (mono wav file format)
func Encode(w io.Writer, s Signal, length interval, sampleRate uint32, sampleBytes uint8) {
	binaryWrite := func(w io.Writer, d interface{}) {
		if err := binary.Write(w, binary.LittleEndian, d); err != nil {
			panic(err.Error()+fmt.Sprint(w,d))
		}
	}
	samplePeriod := MultiplyInterval(1/float32(sampleRate), UnitTime)
	samples := uint32(length/samplePeriod) + 1
	fmt.Fprint(w, "RIFF")
	binaryWrite(w, samples*uint32(sampleBytes)+36)
	fmt.Fprint(w, "WAVE")
	fmt.Fprint(w, "fmt ")
	binaryWrite(w, uint32(16))
	binaryWrite(w, uint16(1))
	binaryWrite(w, uint16(1))
	binaryWrite(w, sampleRate)
	binaryWrite(w, sampleRate*uint32(sampleBytes))
	binaryWrite(w, uint16(sampleBytes))
	binaryWrite(w, uint16(8*sampleBytes))
	fmt.Fprint(w, "data")
	binaryWrite(w, samples*uint32(sampleBytes))
	var i uint32
	switch sampleBytes {
	case 1:
		if pcm,ok:=s.(PCM8bit);ok && pcm.SamplePeriod==samplePeriod && pcm.length==length{
			w.Write(pcm.Data) // TODO can cope with shorter length
		} else {
			for ; i < samples; i++ {
				binaryWrite(w, uint8(s.Level(interval(i)*samplePeriod)>>(LevelBits-8)+128))
			}
		}
	case 2:
		if pcm,ok:=s.(PCM16bit);ok && pcm.SamplePeriod==samplePeriod && pcm.length==length{
			w.Write(pcm.Data) // TODO can cope with shorter length
		} else {
			for ; i < samples; i++ {
				binaryWrite(w, int16(s.Level(interval(i)*samplePeriod)>>(LevelBits-16)))
			}
		}
	case 3:
		buf := bytes.NewBuffer(make([]byte, 4))
		for ; i < samples; i++ {
			binaryWrite(buf, int32(s.Level(interval(i)*samplePeriod)>>(LevelBits-32)))
			w.Write(buf.Bytes()[1:])
		}

	case 4:
		for ; i < samples; i++ {
			binaryWrite(w, int32(s.Level(interval(i)*samplePeriod)>>(LevelBits-32)))
		}
	}
}


type ErrWavParse struct {
	description string
	source      io.Reader
}

func (e ErrWavParse) Error() string {
	switch st := e.source.(type) {
	case *os.File:
		return fmt.Sprintf("WAVE Parse,%s File:%s", e.description, st.Name())
	}
	return fmt.Sprintf("WAVE Parse,%s:%v", e.description, e.source)
}

type limitedSignal interface{
	Signal
	Duration() interval
}

type PCM struct {
	SamplePeriod interval
	length interval
	Data        []uint8
}
func (s PCM) Duration() interval{
	return s.length
}

type PCM8bit struct{
	PCM
	}

func (s PCM8bit) Level(offset interval) level {
	index := int(offset / s.SamplePeriod )
	if index < 0 || index >= len(s.Data)-1 {
		return 0
	}
	return level(s.Data[index]-128) * (MaxLevel >> 7)
}


type PCM16bit struct{
	PCM
	}

func (s PCM16bit) Level(offset interval) level {
	index := int(offset / s.SamplePeriod )*2
	if index < 0 || index >= len(s.Data)-3 {
		return 0
	}
	return level(int16(s.Data[index]) | int16(s.Data[index+1])<<8)* (MaxLevel >> 15)
}

type riffHeader struct {
	C1, C2, C3, C4 byte
	DataLen        uint32
	C5, C6, C7, C8 byte
}

type chunkHeader struct {
	C1, C2, C3, C4 byte
	DataLen        uint32
}

type format struct {
	Code        uint16
	Channels    uint16
	SampleRate  uint32
	ByteRate    uint32
	SampleBytes uint16
	Bits        uint16
}

func Decode(wav io.Reader) ([]limitedSignal, error) {
	var header riffHeader
	var formatHeader chunkHeader
	var formatData format
	var dataHeader chunkHeader
	if err := binary.Read(wav, binary.LittleEndian, &header); err != nil {
		return nil, ErrWavParse{"Not enough data.", wav}
	}
	if header.C1 != 'R' || header.C2 != 'I' || header.C3 != 'F' || header.C4 != 'F' || header.C5 != 'W' || header.C6 != 'A' || header.C7 != 'V' || header.C8 != 'E' {
		return nil, ErrWavParse{"not WAVE format.", wav}
	}
	//var runningBytes int =16
	if err := binary.Read(wav, binary.LittleEndian, &formatHeader); err != nil {
		return nil, ErrWavParse{"Not enough data.", wav}
	}
	if formatHeader.C1 != 'f' || formatHeader.C2 != 'm' || formatHeader.C3 != 't' || formatHeader.C4 != ' ' || formatHeader.DataLen != 16 {
		return nil, ErrWavParse{"no format chunk.", wav}
	}

	if err := binary.Read(wav, binary.LittleEndian, &formatData); err != nil {
		return nil, ErrWavParse{"Not enough data.", wav}
	}
	if formatData.Code != 1 {
		return nil, ErrWavParse{"not PCM format.", wav}
	}
	if formatData.Channels == 0 || formatData.Channels > 2 {
		return nil, ErrWavParse{"not mono or stereo.", wav}
	}
	if formatData.Bits%8 !=0 {
		return nil, ErrWavParse{"not whole byte samples size!", wav}
	}

//	ByteRate    uint32
//	SampleBytes uint16
	


	// need to skip any non-"data" chucks
	if err := binary.Read(wav, binary.LittleEndian, &dataHeader); err != nil {
		return nil, ErrWavParse{"Not enough data.", wav}
	}
	for dataHeader.C1 != 'd' || dataHeader.C2 != 'a' || dataHeader.C3 != 't' || dataHeader.C4 != 'a' {
		var err error
		if s, ok := wav.(io.Seeker); ok {
			_, err = s.Seek(int64(dataHeader.DataLen), os.SEEK_CUR) // seek relative to current file pointer
		} else {
			_, err = io.CopyN(ioutil.Discard, wav, int64(dataHeader.DataLen))
		}
		if err != nil {
			return nil, ErrWavParse{string(dataHeader.C1) + string(dataHeader.C2) + string(dataHeader.C3) + string(dataHeader.C4) + " " + err.Error(), wav}
		}

		if err := binary.Read(wav, binary.LittleEndian, &dataHeader); err != nil {
			return nil, ErrWavParse{"Not enough data.", wav}
		}
	}

	//if dataHeader.DataLen!=header.DataLen-36 {return nil, ErrWavParse{fmt.Sprintf("data chunk size mismatch. %v+36!=%v",dataHeader.DataLen,header.DataLen), []byte(fmt.Sprintf("%#v",dataHeader))}}	//  this is only true for non-extensible wav, ie non-microsoft
	if dataHeader.DataLen%uint32(formatData.Channels) != 0 {
		return nil, ErrWavParse{"sound sample data length not divisable by channel count", wav}
	}

	sampleData := make([]byte, dataHeader.DataLen)

	samples := dataHeader.DataLen / uint32(formatData.Channels)/ uint32(formatData.Bits/8)
	var s uint32
	for ; s < samples; s++ {
		// deinterlace channels by reading directly into consecutive blocks
		var c uint32
		for ; c < uint32(formatData.Channels); c++ {
			if n, err := wav.Read(sampleData[(c*samples+s)*uint32(formatData.Bits/8) : (c*samples+s+1)*uint32(formatData.Bits/8)]); err != nil || n != int(formatData.Bits/8) {
				return nil, ErrWavParse{"read incomplete", wav}
			}

		}
	}
	signals := make([]limitedSignal, formatData.Channels)

	var c uint32
	if formatData.Bits == 8 {
		for ; c < uint32(formatData.Channels); c++ {
			signals[c] = PCM8bit{PCM{UnitTime / interval(formatData.SampleRate) ,UnitTime / interval(formatData.SampleRate) * interval(samples) ,sampleData[c*samples : (c+1)*samples]}}
		}
	}else if formatData.Bits == 16 {
		for ; c < uint32(formatData.Channels); c++ {
			signals[c] = PCM16bit{PCM{UnitTime / interval(formatData.SampleRate) ,UnitTime / interval(formatData.SampleRate) * interval(samples) ,sampleData[c*samples*2 : (c+1)*samples*2]}}
		}

	}
	return signals, nil
}




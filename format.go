package signals

import (
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"errors"
	"encoding/binary"
	"bufio"
)

// RIFF chunk header holder
type chunkHeader struct {
	Id [4]byte
	DataLen        uint32
}

type formatChunk struct {
	Code        uint16
	Channels    uint16
	SampleRate  uint32
	ByteRate    uint32
	SampleBytes uint16
	Bits        uint16
}


// Encode Signals as PCM data,in a Riff wave container.
func Encode(w io.Writer, sampleBytes uint8, sampleRate uint32, length x, ss ...Signal) (err error) {
	buf:=bufio.NewWriter(w)
	err = encode(buf, sampleBytes, sampleRate, length, ss...) 
	if err == nil {	buf.Flush()	}
	return err
}

// unbuffered encode Signals as PCM data,in a Riff wave container.
func encode(w io.Writer, sampleBytes uint8, sampleRate uint32, length x, ss ...Signal) (err error) {
	samplePeriod := X(1 / float32(sampleRate))
	samples := uint32(length/samplePeriod) + 1
	binary.Write(w, binary.LittleEndian, chunkHeader{[4]byte{'R', 'I', 'F', 'F'}, samples*uint32(sampleBytes) + 36})
	w.Write([]byte{'W', 'A', 'V', 'E'})
	binary.Write(w, binary.LittleEndian, chunkHeader{[4]byte{'f', 'm', 't', ' '}, 16})
	binary.Write(w, binary.LittleEndian, formatChunk{
		Code:        1,
		Channels:    uint16(len(ss)),
		SampleRate:  sampleRate,
		ByteRate:    sampleRate * uint32(sampleBytes) * uint32(len(ss)),
		SampleBytes: uint16(sampleBytes) * uint16(len(ss)),
		Bits:        uint16(8 * sampleBytes),
	})
	binary.Write(w, binary.LittleEndian, chunkHeader{[4]byte{'d', 'a', 't', 'a'}, samples * uint32(sampleBytes) * uint32(len(ss))})
	readerForPCM8Bit := func(s Signal) io.Reader {
		r, w := io.Pipe()
		go func() {
			// try short-cuts first
			offset, ok := s.(Offset)
			if pcms, ok2 := offset.LimitedSignal.(PCM8bit); ok && ok2 && pcms.samplePeriod == samplePeriod && pcms.MaxX() >= length-offset.Offset {
				offsetSamples := int(offset.Offset / samplePeriod)
				if offsetSamples < 0 {
					w.Write(pcms.Data[-offsetSamples : -offsetSamples+int(samples)])
				} else {
					for x, zeroSample := 0, []byte{0x80}; x < offsetSamples; x++ {
						w.Write(zeroSample)
					}
					w.Write(pcms.Data[:int(samples)-offsetSamples])
				}
				w.Close()
			} else if pcm, ok := s.(PCM8bit); ok && pcm.samplePeriod == samplePeriod && pcm.MaxX() >= length {
				w.Write(pcm.Data[:samples])
				w.Close()
			} else {
				defer func() {
					e := recover()
					if e != nil {
						w.CloseWithError(e.(error))
					} else {
						w.Close()
					}
				}()
				for i, sample := uint32(0), make([]byte, 1); err == nil && i < samples; i++ {
					sample[0] = encodePCM8bit(s.property(x(i) * samplePeriod))
					_, err = w.Write(sample)
				}
			}
		}()
		return r
	}
	readerForPCM16Bit := func(s Signal) io.Reader {
		r, w := io.Pipe()
		go func() {
			// try shortcuts first
			offset, ok := s.(Offset)
			if pcms, ok2 := offset.LimitedSignal.(PCM16bit); ok && ok2 && pcms.samplePeriod == samplePeriod && pcms.MaxX() >= length-offset.Offset {
				offsetSamples := int(offset.Offset / samplePeriod)
				if offsetSamples < 0 {
					w.Write(pcms.Data[-offsetSamples*2 : (int(samples)-offsetSamples)*2])
				} else {
					for x, zeroSample := 0, []byte{0x00, 0x00}; x < offsetSamples; x++ {
						w.Write(zeroSample)
					}
					w.Write(pcms.Data[:(int(samples)-offsetSamples)*2])
				}
				w.Close()
			} else if pcm, ok := s.(PCM16bit); ok && pcm.samplePeriod == samplePeriod && pcm.MaxX() >= length {
				w.Write(pcm.Data[:samples*2])
				w.Close()
			} else {
				defer func() {
					e := recover()
					if e != nil {
						w.CloseWithError(e.(error))
					} else {
						w.Close()
					}
				}()

				for i, sample := uint32(0), make([]byte, 2); err == nil && i < samples; i++ {
					sample[0], sample[1] = encodePCM16bit(s.property(x(i) * samplePeriod))
					_, err = w.Write(sample)
				}
			}
		}()
		return r
	}
	readerForPCM24Bit := func(s Signal) io.Reader {
		r, w := io.Pipe()
		go func() {
			// try shortcuts first
			offset, ok := s.(Offset)
			if pcms, ok2 := offset.LimitedSignal.(PCM24bit); ok && ok2 && pcms.samplePeriod == samplePeriod && pcms.MaxX() >= length-offset.Offset {
				offsetSamples := int(offset.Offset / samplePeriod)
				if offsetSamples < 0 {
					w.Write(pcms.Data[-offsetSamples*3 : (int(samples)-offsetSamples)*3])
				} else {
					for x, zeroSample := 0, []byte{0x00, 0x00, 0x00}; x < offsetSamples; x++ {
						w.Write(zeroSample)
					}
					w.Write(pcms.Data[:(int(samples)-offsetSamples)*3])
				}
				w.Close()
			} else if pcm, ok := s.(PCM24bit); ok && pcm.samplePeriod == samplePeriod && pcm.MaxX() >= length {
				w.Write(pcm.Data[:samples*3])
				w.Close()
			} else {
				defer func() {
					e := recover()
					if e != nil {
						w.CloseWithError(e.(error))
					} else {
						w.Close()
					}
				}()
				for i, sample := uint32(0), make([]byte, 3); err == nil && i < samples; i++ {
					sample[0], sample[1], sample[2] = encodePCM24bit(s.property(x(i) * samplePeriod))
					_, err = w.Write(sample)
				}
			}
		}()
		return r
	}
	readerForPCM32Bit := func(s Signal) io.Reader {
		r, w := io.Pipe()
		go func() {
			// try shortcuts first
			offset, ok := s.(Offset)
			if pcms, ok2 := offset.LimitedSignal.(PCM32bit); ok && ok2 && pcms.samplePeriod == samplePeriod && pcms.MaxX() >= length-offset.Offset {
				offsetSamples := int(offset.Offset / samplePeriod)
				if offsetSamples < 0 {
					w.Write(pcms.Data[-offsetSamples*4 : (int(samples)-offsetSamples)*4])
				} else {
					for x, zeroSample := 0, []byte{0x00, 0x00, 0x00, 0x00}; x < offsetSamples; x++ {
						w.Write(zeroSample)
					}
					w.Write(pcms.Data[:(int(samples)-offsetSamples)*4])
				}
				w.Close()
			} else if pcm, ok := s.(PCM32bit); ok && pcm.samplePeriod == samplePeriod && pcm.MaxX() >= length {
				w.Write(pcm.Data[:samples*4])
				w.Close()
			} else {
				defer func() {
					e := recover()
					if e != nil {
						w.CloseWithError(e.(error))
					} else {
						w.Close()
					}
				}()
				for i, sample := uint32(0), make([]byte, 4); err == nil && i < samples; i++ {
					sample[0], sample[1], sample[2], sample[3] = encodePCM32bit(s.property(x(i) * samplePeriod))
					_, err = w.Write(sample)
				}
			}
		}()
		return r
	}
	readerForPCM48Bit := func(s Signal) io.Reader {
		r, w := io.Pipe()
		go func() {
			// try shortcuts first
			offset, ok := s.(Offset)
			if pcms, ok2 := offset.LimitedSignal.(PCM48bit); ok && ok2 && pcms.samplePeriod == samplePeriod && pcms.MaxX() >= length-offset.Offset {
				offsetSamples := int(offset.Offset / samplePeriod)
				if offsetSamples < 0 {
					w.Write(pcms.Data[-offsetSamples*6 : (int(samples)-offsetSamples)*6])
				} else {
					for x, zeroSample := 0, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}; x < offsetSamples; x++ {
						w.Write(zeroSample)
					}
					w.Write(pcms.Data[:(int(samples)-offsetSamples)*6])
				}
				w.Close()
			} else if pcm, ok := s.(PCM48bit); ok && pcm.samplePeriod == samplePeriod && pcm.MaxX() >= length {
				w.Write(pcm.Data[:samples*6])
				w.Close()
			} else {
				defer func() {
					e := recover()
					if e != nil {
						w.CloseWithError(e.(error))
					} else {
						w.Close()
					}
				}()
				for i, sample := uint32(0), make([]byte, 6); err == nil && i < samples; i++ {
					sample[0], sample[1], sample[2], sample[3], sample[4], sample[5] = encodePCM48bit(s.property(x(i) * samplePeriod))
					_, err = w.Write(sample)
				}
			}
		}()
		return r
	}
	readerForPCM64Bit := func(s Signal) io.Reader {
		r, w := io.Pipe()
		go func() {
			// try shortcuts first
			offset, ok := s.(Offset)
			if pcms, ok2 := offset.LimitedSignal.(PCM64bit); ok && ok2 && pcms.samplePeriod == samplePeriod && pcms.MaxX() >= length-offset.Offset {
				offsetSamples := int(offset.Offset / samplePeriod)
				if offsetSamples < 0 {
					w.Write(pcms.Data[-offsetSamples*8 : (int(samples)-offsetSamples)*8])
				} else {
					for x, zeroSample := 0, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}; x < offsetSamples; x++ {
						w.Write(zeroSample)
					}
					w.Write(pcms.Data[:(int(samples)-offsetSamples)*8])
				}
				w.Close()
			} else if pcm, ok := s.(PCM64bit); ok && pcm.samplePeriod == samplePeriod && pcm.MaxX() >= length {
				w.Write(pcm.Data[:samples*8])
				w.Close()
			} else {
				defer func() {
					e := recover()
					if e != nil {
						w.CloseWithError(e.(error))
					} else {
						w.Close()
					}
				}()
				for i, sample := uint32(0), make([]byte, 8); err == nil && i < samples; i++ {
					sample[0], sample[1], sample[2], sample[3], sample[4], sample[5], sample[6], sample[7] = encodePCM64bit(s.property(x(i) * samplePeriod))
					_, err = w.Write(sample)
				}
			}
		}()
		return r
	}
	readers := make([]io.Reader, len(ss))
	switch sampleBytes {
	case 1:
		for i, _ := range readers {
			readers[i] = readerForPCM8Bit(ss[i])
		}
		err = interleavedWrite(w, 1, readers...)
	case 2:
		for i, _ := range readers {
			readers[i] = readerForPCM16Bit(ss[i])
		}
		err = interleavedWrite(w, 2, readers...)
	case 3:
		for i, _ := range readers {
			readers[i] = readerForPCM24Bit(ss[i])
		}
		err = interleavedWrite(w, 3, readers...)
	case 4:
		for i, _ := range readers {
			readers[i] = readerForPCM32Bit(ss[i])
		}
		err = interleavedWrite(w, 4, readers...)
	case 6:
		for i, _ := range readers {
			readers[i] = readerForPCM48Bit(ss[i])
		}
		err = interleavedWrite(w, 6, readers...)
	case 8:
		for i, _ := range readers {
			readers[i] = readerForPCM64Bit(ss[i])
		}
		err = interleavedWrite(w, 8, readers...)
	}
	return
}

func interleavedWrite(w io.Writer, blockSize int64, rs ...io.Reader) (err error) {
	if len(rs) == 0 {
		return
	}
	if len(rs) == 1 {
		_, err = io.Copy(w, rs[0])
	} else {
		for err == nil {
			for i, _ := range rs {
				_, err = io.CopyN(w, rs[i], blockSize)
			}
		}
		if err == io.EOF {
			err = nil
		}
	}
	return
}

// encode a LimitedSignal with a sampleRate equal to the Period() of a given PeriodicSignal, and its precision if its a PCM type, otherwise defaults to 16bit.
func EncodeLike(w io.Writer, s PeriodicSignal, p LimitedSignal) {
	switch s.(type) {
	case PCM8bit:
		Encode(w, 1, uint32(unitX/s.Period()), p.MaxX(), p)
	case PCM16bit:
		Encode(w, 2, uint32(unitX/s.Period()), p.MaxX(), p)
	case PCM24bit:
		Encode(w, 3, uint32(unitX/s.Period()), p.MaxX(), p)
	case PCM32bit:
		Encode(w, 4, uint32(unitX/s.Period()), p.MaxX(), p)
	case PCM48bit:
		Encode(w, 6, uint32(unitX/s.Period()), p.MaxX(), p)
	case PCM64bit:
		Encode(w, 8, uint32(unitX/s.Period()), p.MaxX(), p)
	default:
		Encode(w, 2, uint32(unitX/s.Period()), p.MaxX(), p)
	}
	return
}


type errParsing struct {
	error
	r io.Reader
}

func (e errParsing) Parsing() io.Reader {
	return e.r
}

// Read a wave format stream into an array of PeriodicLimitedSignals.
// one for each channel in the encoding.
func Decode(wav io.Reader) ([]PeriodicLimitedSignal, error) {
	bytesToRead, format, err := readWaveHeader(wav)
	if err != nil {
		return nil, err
	}
	samples := bytesToRead / uint32(format.Channels) / uint32(format.Bits/8)
	sampleData, err := readInterleaved(wav, samples, uint32(format.Channels), uint32(format.Bits/8))
	if err != nil {
		return nil, err
	}
	pcms := make([]PeriodicLimitedSignal, format.Channels)
	switch format.Bits {
	case 8:
		for c := uint32(0); c < uint32(format.Channels); c++ {
			pcms[c] = PCM8bit{PCM{unitX / x(format.SampleRate), sampleData[c*samples : (c+1)*samples]}}
		}
	case 16:
		for c := uint32(0); c < uint32(format.Channels); c++ {
			pcms[c] = PCM16bit{PCM{unitX / x(format.SampleRate), sampleData[c*samples*2 : (c+1)*samples*2]}}
		}
	case 24:
		for c := uint32(0); c < uint32(format.Channels); c++ {
			pcms[c] = PCM24bit{PCM{unitX / x(format.SampleRate), sampleData[c*samples*3 : (c+1)*samples*3]}}
		}
	case 32:
		for c := uint32(0); c < uint32(format.Channels); c++ {
			pcms[c] = PCM32bit{PCM{unitX / x(format.SampleRate), sampleData[c*samples*4 : (c+1)*samples*4]}}
		}
	case 48:
		for c := uint32(0); c < uint32(format.Channels); c++ {
			pcms[c] = PCM48bit{PCM{unitX / x(format.SampleRate), sampleData[c*samples*6 : (c+1)*samples*6]}}
		}
	case 64:
		for c := uint32(0); c < uint32(format.Channels); c++ {
			pcms[c] = PCM64bit{PCM{unitX / x(format.SampleRate), sampleData[c*samples*8 : (c+1)*samples*8]}}
		}
	default:
		return nil,errParsing{errors.New(fmt.Sprintf("Unsupported bit depth (%d).", format.Bits)),wav}
	}
	return pcms, nil
}

func readWaveHeader(wav io.Reader) (uint32, *formatChunk, error) {
	var header chunkHeader
	var formatHeader chunkHeader
	var format formatChunk
	var dataHeader chunkHeader
	if err := binary.Read(wav, binary.LittleEndian, &header); err != nil {
		return 0, nil, errParsing{err,wav}
	}
	if header.Id != [4]byte{'R','I','F','F'}{
		return 0, nil, errParsing{errors.New("Not RIFF format."),wav}
	}
	b:=make([]byte,4)
	if _,err:=wav.Read(b); err != nil{
		return 0, nil, errParsing{err,wav}
	}
	if b[0] != 'W' || b[1] != 'A' || b[2] != 'V' || b[3] != 'E' {
		return 0, nil, errParsing{errors.New("Not WAVE format."),wav}
	} 
	if err := binary.Read(wav, binary.LittleEndian, &formatHeader); err != nil {
		return 0, nil, errParsing{err,wav}
	}
	// skip any non-"fmt " chunks
	for formatHeader.Id != [4]byte{'f','m','t',' '} {
		var err error
		if s, ok := wav.(io.Seeker); ok {
			_, err = s.Seek(int64(formatHeader.DataLen), os.SEEK_CUR) // seek relative to current file pointer if possible
		} else {
			_, err = io.CopyN(ioutil.Discard, wav, int64(formatHeader.DataLen))
		}
		if err != nil {
			return 0, &format, errParsing{errors.New(fmt.Sprint(formatHeader.Id) + " " + err.Error()),wav}
		}

		if err := binary.Read(wav, binary.LittleEndian, &formatHeader); err != nil {
			return 0, &format, errParsing{err,wav}
		}
	}

	if formatHeader.DataLen != 16 {
		return 0, nil, errParsing{errors.New("Format chunk wrong size." + string(formatHeader.DataLen)),wav}
	}

	if err := binary.Read(wav, binary.LittleEndian, &format); err != nil {
		return 0, nil, errParsing{err,wav}
	}
	if format.Code != 1 {
		return 0, &format, errParsing{errors.New("only PCM supported. not format code:" + string(format.Code)),wav}
	}
	if format.Bits%8 != 0 {
		return 0, &format, errParsing{errors.New("not whole byte samples size!"),wav}
	}

	// TODO-nice read "LIST" chunk with, 3 fields, third being "INFO", can contain "ICOP" and "ICRD" chunks providing copyright and creation date information.

	// skip any non-"data" chucks
	if err := binary.Read(wav, binary.LittleEndian, &dataHeader); err != nil {
		return 0, &format, errParsing{err,wav}
	}
	for dataHeader.Id[0] != 'd' || dataHeader.Id[1] != 'a' || dataHeader.Id[2] != 't' || dataHeader.Id[3] != 'a' {
		var err error
		if s, ok := wav.(io.Seeker); ok {
			_, err = s.Seek(int64(dataHeader.DataLen), os.SEEK_CUR) // seek relative to current file pointer if possible
		} else {
			_, err = io.CopyN(ioutil.Discard, wav, int64(dataHeader.DataLen))
		}
		if err != nil {
			return 0, &format, errParsing{errors.New(fmt.Sprint(formatHeader.Id) + " " + err.Error()),wav}
		}

		if err := binary.Read(wav, binary.LittleEndian, &dataHeader); err != nil {
			return 0, &format, errParsing{err,wav}
		}
	}
	if dataHeader.DataLen%uint32(format.Channels) != 0 {
		return 0, &format, errParsing{errors.New("sound sample data length not divisible by channel count:" + string(dataHeader.DataLen)),wav}
	}
	return dataHeader.DataLen, &format, nil
}

func readInterleaved(r io.Reader, samples uint32, channels uint32, sampleBytes uint32) ([]byte, error) {
	sampleData := make([]byte, samples*channels*sampleBytes)
	var err error
	for s := uint32(0); s < samples; s++ {
		// de-interlace channels by reading directly into separate regions of a byte slice
		// TODO better to de-interlace into separate readers?
		for c := uint32(0); c < uint32(channels); c++ {
			if n, err := r.Read(sampleData[(c*samples+s)*sampleBytes : (c*samples+s+1)*sampleBytes]); err != nil || n != int(sampleBytes) {
				return nil, errors.New(fmt.Sprintf("data ran out at %v of %v", s, samples))
			}
		}
	}
	return sampleData, err
}



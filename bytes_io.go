// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2020
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE. }}}

package sdr

import (
	"encoding/binary"
	"io"

	"hz.tools/sdr/internal"
)

type byteWriterForeign struct {
	w                io.Writer
	byteOrder        binary.ByteOrder
	samplesPerSecond uint
	sampleFormat     SampleFormat
}

func (bw byteWriterForeign) Write(samples Samples) (int, error) {
	if samples.Format() != bw.sampleFormat {
		return 0, ErrSampleFormatMismatch
	}

	switch buf := samples.(type) {
	case SamplesU8:
		bufBytes, err := UnsafeSamplesAsBytes(buf)
		if err != nil {
			return 0, err
		}
		i, err := bw.w.Write(bufBytes)
		return i / 2, err
	case SamplesI16:
		if err := binary.Write(bw.w, bw.byteOrder, buf); err != nil {
			return 0, err
		}
		return len(buf), nil
	case SamplesC64:
		if err := binary.Write(bw.w, bw.byteOrder, buf); err != nil {
			return 0, err
		}
		return len(buf), nil
	default:
		return 0, ErrSampleFormatUnknown
	}
}

func (bw byteWriterForeign) SampleRate() uint {
	return bw.samplesPerSecond
}

func (bw byteWriterForeign) SampleFormat() SampleFormat {
	return bw.sampleFormat
}

type byteWriterNative struct {
	w                io.Writer
	samplesPerSecond uint
	sampleFormat     SampleFormat
}

func (bw byteWriterNative) Write(samples Samples) (int, error) {
	if samples.Format() != bw.sampleFormat {
		return 0, ErrSampleFormatMismatch
	}

	switch buf := samples.(type) {
	case SamplesU8:
		bufBytes, err := UnsafeSamplesAsBytes(buf)
		if err != nil {
			return 0, err
		}
		i, err := bw.w.Write(bufBytes)
		return i / SampleFormatU8.Size(), err
	case SamplesI16:
		bufBytes, err := UnsafeSamplesAsBytes(buf)
		if err != nil {
			return 0, err
		}
		i, err := bw.w.Write(bufBytes)
		return i / SampleFormatI16.Size(), err
	case SamplesC64:
		bufBytes, err := UnsafeSamplesAsBytes(buf)
		if err != nil {
			return 0, err
		}
		i, err := bw.w.Write(bufBytes)
		return i / SampleFormatC64.Size(), err
	default:
		return 0, ErrSampleFormatUnknown
	}
}

func (bw byteWriterNative) SampleRate() uint {
	return bw.samplesPerSecond
}

func (bw byteWriterNative) SampleFormat() SampleFormat {
	return bw.sampleFormat
}

// ByteWriter will wrap an io.Writer, and write encoded IQ data as a series
// of raw bytes out.
func ByteWriter(
	w io.Writer,
	byteOrder binary.ByteOrder,
	samplesPerSecond uint,
	sf SampleFormat,
) Writer {
	if byteOrder == internal.NativeEndian {
		return byteWriterNative{
			w:                w,
			samplesPerSecond: samplesPerSecond,
			sampleFormat:     sf,
		}
	}

	return byteWriterForeign{
		w:                w,
		byteOrder:        byteOrder,
		samplesPerSecond: samplesPerSecond,
		sampleFormat:     sf,
	}
}

type byteReaderForeign struct {
	r                io.Reader
	byteOrder        binary.ByteOrder
	samplesPerSecond uint
	sampleFormat     SampleFormat
}

func (br byteReaderForeign) Read(samples Samples) (int, error) {
	if samples.Format() != br.sampleFormat {
		return 0, ErrSampleFormatMismatch
	}

	switch buf := samples.(type) {
	case SamplesU8:
		bufBytes, err := UnsafeSamplesAsBytes(buf)
		if err != nil {
			return 0, err
		}
		i, err := br.r.Read(bufBytes)
		return i / SampleFormatU8.Size(), err
	case SamplesI16:
		// TODO(paultag): binary.Read here forces a ReadFull which isn't
		// ideal.
		err := binary.Read(br.r, br.byteOrder, buf)
		return buf.Length(), err
	case SamplesC64:
		err := binary.Read(br.r, br.byteOrder, buf)
		return buf.Length(), err
	default:
		return 0, ErrSampleFormatUnknown
	}
}

func (br byteReaderForeign) SampleFormat() SampleFormat {
	return br.sampleFormat
}

func (br byteReaderForeign) SampleRate() uint {
	return br.samplesPerSecond
}

type byteReaderNative struct {
	r                io.Reader
	samplesPerSecond uint
	sampleFormat     SampleFormat
}

func (br byteReaderNative) Read(samples Samples) (int, error) {
	if samples.Format() != br.sampleFormat {
		return 0, ErrSampleFormatMismatch
	}

	switch buf := samples.(type) {
	case SamplesU8:
		bufBytes, err := UnsafeSamplesAsBytes(buf)
		if err != nil {
			return 0, err
		}
		i, err := br.r.Read(bufBytes)
		return i / SampleFormatU8.Size(), err
	case SamplesI16:
		bufBytes, err := UnsafeSamplesAsBytes(buf)
		if err != nil {
			return 0, err
		}
		i, err := br.r.Read(bufBytes)
		return i / SampleFormatI16.Size(), err
	case SamplesC64:
		bufBytes, err := UnsafeSamplesAsBytes(buf)
		if err != nil {
			return 0, err
		}
		i, err := br.r.Read(bufBytes)
		return i / SampleFormatC64.Size(), err
	default:
		return 0, ErrSampleFormatUnknown
	}
}

func (br byteReaderNative) SampleFormat() SampleFormat {
	return br.sampleFormat
}

func (br byteReaderNative) SampleRate() uint {
	return br.samplesPerSecond
}

// ByteReader will wrap an io.Reader, and decode encoded IQ data from raw
// bytes into an sdr.Samples object.
func ByteReader(
	r io.Reader,
	byteOrder binary.ByteOrder,
	samplesPerSecond uint,
	sf SampleFormat,
) Reader {
	if byteOrder == internal.NativeEndian {
		return byteReaderNative{
			r:                r,
			samplesPerSecond: samplesPerSecond,
			sampleFormat:     sf,
		}
	}

	return byteReaderForeign{
		r:                r,
		byteOrder:        byteOrder,
		samplesPerSecond: samplesPerSecond,
		sampleFormat:     sf,
	}
}

// vim: foldmethod=marker

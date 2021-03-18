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

package stream

import (
	"fmt"

	"hz.tools/sdr"
)

// ConvertReader will convert one reader in a specific SampleFormat to a
// reader of another format.
//
// This can be used is code is expecting to be reading from a sdr.Reader and
// requires that the IQ samples be in a specific format. Practically, this
// lets the code read from an rtl-sdr or hackrf in uint8 format, and get
// samples in complex64, or vice-versa (read from a complex64 capture and
// process as if it was an rtl-sdr).
func ConvertReader(in sdr.Reader, to sdr.SampleFormat) (sdr.Reader, error) {
	return ReadTransformer(in, ReadTransformerConfig{
		InputBufferLength:  32 * 1024,
		OutputBufferLength: 32 * 1024,
		OutputSampleRate:   in.SampleRate(),
		OutputSampleFormat: to,
		Proc: func(inBuf sdr.Samples, outBuf sdr.Samples) (int, error) {
			return sdr.ConvertBuffer(outBuf, inBuf)
		},
	})
}

// ConvertWriter will convert one writer of a specific type to a writer
// of another SampleFormat.
//
// ConvertWriter will take the output Writer (out) and create a new writer
// with a new sample format (inputFormat).
func ConvertWriter(
	out sdr.Writer,
	inputFormat sdr.SampleFormat,
) (sdr.Writer, error) {
	bufSize := 32 * 1024
	buf, err := sdr.MakeSamples(out.SampleFormat(), bufSize)
	if err != nil {
		return nil, err
	}

	return &convWriter{
		out:         out,
		inputFormat: inputFormat,
		buffer:      buf,
	}, nil
}

type convWriter struct {
	out         sdr.Writer
	inputFormat sdr.SampleFormat
	buffer      sdr.Samples
}

func (cw convWriter) Write(in sdr.Samples) (int, error) {
	if in.Format() != cw.inputFormat {
		return 0, sdr.ErrSampleFormatMismatch
	}

	bufSize := cw.buffer.Length()

	n := 0

	for i := 0; i < in.Length(); i += bufSize {
		ie := i + bufSize
		if ie > in.Length() {
			ie = in.Length()
		}

		leng, err := sdr.ConvertBuffer(cw.buffer, in.Slice(i, ie))
		if err != nil {
			return n, err
		}

		if ie-i != leng {
			return n, fmt.Errorf("ConvertWriter: Conversion mismatch")
		}

		j, err := cw.out.Write(cw.buffer.Slice(0, leng))

		n += j
		if err != nil {
			return n, err
		}
	}

	return n, nil
}

func (cw convWriter) SampleFormat() sdr.SampleFormat {
	return cw.inputFormat
}

func (cw convWriter) SampleRate() uint {
	return cw.out.SampleRate()
}

// vim: foldmethod=marker

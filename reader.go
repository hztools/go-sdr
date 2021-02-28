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
	"fmt"
	"io"
)

var (
	// ErrShortBuffer will return if the number of bytes read was less than the
	// minimum required by the callee.
	ErrShortBuffer error = fmt.Errorf("sdr: short read")

	// ErrUnexpectedEOF will return if the EOF was reached before parsing was
	// completed.
	ErrUnexpectedEOF error = fmt.Errorf("sdr: expected EOF")
)

// Reader is the interface that wraps the basic Read method.
type Reader interface {
	// Read IQ Samples into the target Samples buffer. There are two return
	// values, an int representing the **IQ** samples (not bytes) read by this
	// function, and any error conditions encountered.
	Read(Samples) (int, error)

	// Get the sdr.SampleFormat
	SampleFormat() SampleFormat

	// SampleRate will get the number of samples per second that this
	// stream is communicating at.
	SampleRate() uint32
}

// Closer is the interface that wraps the basic Close method.
type Closer interface {
	Close() error
}

// ReadCloser is the interface that groups the basic Read and Close methods.
type ReadCloser interface {
	Reader
	Closer
}

// ReadFull reads exactly len(buf) bytes from r into buf.
func ReadFull(r Reader, buf Samples) (int, error) {
	return ReadAtLeast(r, buf, buf.Length())
}

type readerWithCloser struct {
	Reader
	closer func() error
}

func (rwc readerWithCloser) Close() error {
	return rwc.closer()
}

// ReaderWithCloser will add a closer to a reader to make an sdr.ReadCloser
func ReaderWithCloser(r Reader, c func() error) ReadCloser {
	return readerWithCloser{
		Reader: r,
		closer: c,
	}
}

// ReadAtLeast reads from r into buf until it has read at least min bytes.
func ReadAtLeast(r Reader, buf Samples, min int) (int, error) {
	if buf.Length() < min {
		return 0, ErrShortBuffer
	}
	var (
		n   int
		err error
	)
	for n < min && err == nil {
		var nn int
		nn, err = r.Read(buf.Slice(n, buf.Length()))
		n += nn
	}
	if n >= min {
		return n, err
	} else if n > 0 && err == io.EOF {
		return n, ErrUnexpectedEOF
	}
	return n, err
}

type multiReader struct {
	readers      []Reader
	idx          int
	err          error
	sampleFormat SampleFormat
	sampleRate   uint32
}

func (mr *multiReader) Read(s Samples) (int, error) {
	if mr.err != nil {
		return 0, mr.err
	}
	i, err := mr.readers[mr.idx].Read(s)
	if err == io.EOF {
		if mr.idx >= len(mr.readers) {
			mr.err = io.EOF
			return i, err
		}
		mr.idx++
		return i, nil
	}

	if err != nil {
		mr.err = err
	}
	return i, err
}

func (mr *multiReader) SampleFormat() SampleFormat {
	return mr.sampleFormat
}

func (mr *multiReader) SampleRate() uint32 {
	return mr.sampleRate
}

// MultiReader will act like `cat`, passing Reads through from one reader
// to the next until the end of the streams.
//
// An io.EOF will be returned if they all return EOF, otherwise the first error
// to be hit will be returned.
func MultiReader(readers ...Reader) (Reader, error) {
	switch len(readers) {
	case 0:
		return nil, fmt.Errorf("hz.tools/sdr.MultiReader: Must have at least one reader")
	case 1:
		return readers[0], nil
	}

	var (
		sampleFormat SampleFormat = readers[0].SampleFormat()
		sampleRate   uint32       = readers[0].SampleRate()
	)

	for _, reader := range readers[1:] {
		if reader.SampleFormat() != sampleFormat {
			return nil, ErrSampleFormatMismatch
		}
		if reader.SampleRate() != sampleRate {
			return nil, fmt.Errorf("hz.tools/sdr.MultiReader: Sample rate mismatch")
		}
	}

	return &multiReader{
		readers:      readers,
		idx:          0,
		err:          nil,
		sampleFormat: sampleFormat,
		sampleRate:   sampleRate,
	}, nil
}

// vim: foldmethod=marker

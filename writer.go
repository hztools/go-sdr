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
	"errors"
)

// ErrShortWrite will be returned when a write was aborted halfway through.
var ErrShortWrite = errors.New("sdr: short write")

// Writer is the interface that wraps the basic Write method.
type Writer interface {
	// Write IQ Samples into the target Samples buffer. There are two return
	// values, an int representing the **IQ** samples (not bytes) read by this
	// function, and any error conditions encountered.
	Write(Samples) (int, error)

	// Get the sdr.SampleFormat
	SampleFormat() SampleFormat

	// SampleRate will get the number of samples per second that this
	// stream is communicating at.
	SampleRate() uint32
}

// WriteCloser is the interface that groups the basic Read and Close methods.
type WriteCloser interface {
	Writer
	Closer
}

type multiWriter struct {
	writers          []Writer
	samplesPerSecond uint32
	sampleFormat     SampleFormat
}

// MultiWriter creates a writer that duplicates its writes to all the
// provided writers, similar to the Unix tee(1) command, or io.MultiWriter.
//
// This has the same behavior as an io.MultiWriter, but will copy between
// IQ streams.
func MultiWriter(
	samplesPerSecond uint32,
	sampleFormat SampleFormat,
	writers ...Writer,
) Writer {
	allWriters := make([]Writer, 0, len(writers))
	for _, w := range writers {
		if mw, ok := w.(*multiWriter); ok {
			allWriters = append(allWriters, mw.writers...)
		} else {
			allWriters = append(allWriters, w)
		}
	}
	return &multiWriter{
		sampleFormat:     sampleFormat,
		samplesPerSecond: samplesPerSecond,
		writers:          allWriters,
	}
}

func (mw *multiWriter) SampleRate() uint32 {
	return mw.samplesPerSecond
}

// SampleFormat implements the sdr.Writer interface.
func (mw *multiWriter) SampleFormat() SampleFormat {
	return mw.sampleFormat
}

// Write implements the sdr.Writer interface.
func (mw *multiWriter) Write(buf Samples) (int, error) {
	var (
		i   int
		err error
	)

	for _, w := range mw.writers {
		i, err = w.Write(buf)
		if err != nil {
			return i, err
		}
		if i != buf.Length() {
			return i, ErrShortWrite
		}
	}
	return buf.Length(), nil
}

// vim: foldmethod=marker

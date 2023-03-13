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
	"context"
	"fmt"
)

var (
	// ErrPipeClosed will be returned when the Pipe is closed.
	ErrPipeClosed = fmt.Errorf("sdr: pipe is closed")
)

// PipeReader is the Read interface exposed by the Pipe.
type PipeReader interface {
	ReadCloser

	// CloseWithError will Close the pipe with a specific Error rather than
	// the default ErrPipeClosed. This can be useful if code is expecting an
	// io.EOF, for instance.
	CloseWithError(error) error
}

// PipeWriter is the Write interface exposed by the Pipe.
type PipeWriter interface {
	WriteCloser

	// CloseWithError will Close the pipe with a specific Error rather than
	// the default ErrPipeClosed. This can be useful if code is expecting an
	// io.EOF, for instance.
	CloseWithError(error) error
}

// PipeReadWriter is the Read/Write interface exposed by a Pipe.
type PipeReadWriter interface {
	PipeReader
	PipeWriter
}

// pipe is a riff on io.Pipe in the Go stdlib, but tweaked a bit to be
// used with sdr.Samples.
type pipe struct {
	// samples is used to send hunks to be copied to the reader from the writer.

	context context.Context
	cancel  context.CancelFunc

	samplesCh     chan Samples
	readSamplesCh chan int

	samplesPerSecond uint
	format           SampleFormat

	err error
}

// Read implements the sdr.Reader interface.
func (pipe *pipe) Read(b Samples) (int, error) {
	if err := pipe.getErr(); err != nil {
		return 0, err
	}

	if b.Format() != pipe.format {
		return 0, ErrSampleFormatMismatch
	}

	select {
	case sample := <-pipe.samplesCh:
		numRead, err := CopySamples(b, sample)
		pipe.readSamplesCh <- numRead
		return numRead, err
	case <-pipe.context.Done():
		return 0, pipe.getErr()
	}
}

func (pipe *pipe) getErr() error {
	if err := pipe.context.Err(); err == nil {
		return nil
	}
	if pipe.err != nil {
		return pipe.err
	}
	return ErrPipeClosed
}

// Write implements the sdr.Writer interface.
func (pipe *pipe) Write(b Samples) (int, error) {
	if err := pipe.getErr(); err != nil {
		return 0, err
	}

	// TODO(paultag): Thread safety.

	if b.Format() != pipe.format {
		return 0, ErrSampleFormatMismatch
	}

	var n int

	for b.Length() > 0 {
		select {
		case pipe.samplesCh <- b:
			numWritten := <-pipe.readSamplesCh
			b = b.Slice(numWritten, b.Length())
			n += numWritten
		case <-pipe.context.Done():
			return n, pipe.getErr()
		}
	}

	return n, nil
}

// SampleFormat implements the sdr.Reader / sdr.Writer interface.
func (pipe *pipe) SampleRate() uint {
	return pipe.samplesPerSecond
}

// SampleFormat implements the sdr.Reader / sdr.Writer interface.
func (pipe *pipe) SampleFormat() SampleFormat {
	return pipe.format
}

// Close implements the sdr.ReadCloser/sdr.WriteCloser interface.
func (pipe *pipe) CloseWithError(err error) error {
	pipe.err = err
	return pipe.Close()
}

// Close implements the sdr.ReadCloser/sdr.WriteCloser interface.
func (pipe *pipe) Close() error {
	// This should explicitly be not doing anything further, since the core
	// mechanism here is that the context is cancelled / timed out, so relying
	// on this method being called is not a safe assumption. This is merely
	// to adapt the context into a Read/Write Closer to maintain interop
	// with people's mental models and in cases where a context is not passed
	// into the Pipe.
	pipe.cancel()
	return nil
}

// Pipe will create a new sdr.Reader and sdr.Writer that will allow writes
// to pass through and show up to a reader. This allows "patching" a Write
// endpoint into a "Read" endpoint.
func Pipe(samplesPerSecond uint, format SampleFormat) (PipeReader, PipeWriter) {
	ctx := context.Background()
	return PipeWithContext(ctx, samplesPerSecond, format)
}

// PipeWithContext will create a new sdr.Reader and sdr.Writer as returned by
// the Pipe call, but with a custom Context. This is purely used for external
// control of the lifecycle of the Pipe.
func PipeWithContext(
	ctx context.Context,
	samplesPerSecond uint,
	format SampleFormat,
) (PipeReader, PipeWriter) {
	ctx, cancel := context.WithCancel(ctx)
	p := &pipe{
		context:          ctx,
		cancel:           cancel,
		format:           format,
		samplesPerSecond: samplesPerSecond,
		samplesCh:        make(chan Samples),
		readSamplesCh:    make(chan int),
	}
	return p, p
}

// vim: foldmethod=marker

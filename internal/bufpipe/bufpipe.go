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

package bufpipe

import (
	"context"
	"fmt"

	"hz.tools/sdr"
)

var (
	// ErrBufferOverrun will be returned if a write is triggered when there
	// is no remaining capacity.
	ErrBufferOverrun error = fmt.Errorf("sdr/internal/bufpipe: Buffer Overrun")
)

// Dupe will duplicate (and copy) samples from one buffer s1, to a new
// buffer that is freshly allocated, and returned.
func Dupe(s1 sdr.Samples) (sdr.Samples, int, error) {
	s2, err := sdr.MakeSamples(s1.Format(), s1.Length())
	if err != nil {
		return nil, 0, err
	}
	n, err := sdr.CopySamples(s2, s1)
	return s2, n, err
}

// Pipe wraps a normal sdr.Pipe, but writes will not block. Writes are queued
// into a channel.
//
// If the writes exceed the buf capcity, all future writes and reads will return
// an ErrBufferOverrun.
type Pipe struct {
	ctx    context.Context
	cancel context.CancelFunc
	err    error
	buf    chan sdr.Samples

	sampleRate   uint
	sampleFormat sdr.SampleFormat

	pipeReader sdr.PipeReader
	pipeWriter sdr.PipeWriter
}

// SampleFormat implements the sdr.ReadWriter interface.
func (p *Pipe) SampleFormat() sdr.SampleFormat {
	return p.sampleFormat
}

// SampleRate implements the sdr.ReadWriter interface.
func (p *Pipe) SampleRate() uint {
	return p.sampleRate
}

// Read implements the sdr.ReadWriter interface.
func (p *Pipe) Read(s sdr.Samples) (int, error) {
	if p.err != nil {
		// TODO(paultag): Should this exhaust the queue until the end,
		// and then return that? Not sure.
		return 0, p.err
	}

	if s.Format() != p.sampleFormat {
		return 0, sdr.ErrSampleFormatMismatch
	}

	return p.pipeReader.Read(s)
}

func (p *Pipe) do() {
	for {
		select {
		case s1 := <-p.buf:
			_, err := p.pipeWriter.Write(s1)
			if err != nil {
				// If we caught an error and we don't have an error
				// ourselves (like a context error), we're going to
				// go ahead and set the error condition and bail.
				//
				// TODO(paultag): This is not threadsafe.
				if p.err == nil {
					p.err = err
					p.CloseWithError(err)
				}
				return
			}
		case <-p.ctx.Done():
			return
		}
	}
}

// Write implements the sdr.Writer interface.
func (p *Pipe) Write(s1 sdr.Samples) (int, error) {
	if p.err != nil {
		return 0, p.err
	}

	if s1.Format() != p.sampleFormat {
		return 0, sdr.ErrSampleFormatMismatch
	}

	// TODO(paultag): Dupe may not be needed?
	s2, i, err := Dupe(s1)
	if err != nil {
		return 0, err
	}

	select {
	case p.buf <- s2:
		return i, nil
	default:
		p.CloseWithError(ErrBufferOverrun)
		return 0, ErrBufferOverrun
	}
}

// CloseWithError will close the pipe, and return the provided error on any
// subsequent call.
func (p *Pipe) CloseWithError(err error) error {
	p.err = err
	p.Close()
	return nil
}

// Close will cancel the Pipe's context, terminating the goroutine spawned,
// and closing the context of any children objects.
func (p *Pipe) Close() error {
	p.cancel()
	return nil
}

// New will create a new bufpipe.Pipe, which wraps a normal sdr.Pipe,
// but writes will not block.
func New(capacity int, sampleRate uint, sampleFormat sdr.SampleFormat) (*Pipe, error) {
	return NewWithContext(context.Background(), capacity, sampleRate, sampleFormat)
}

// NewWithContext will create a new bufpipe.Pipe, which wraps a normal sdr.Pipe,
// but writes will not block.
//
// This includes a parent context, which if it is expired or is cancelled
// will trigger a close of this context as well.
func NewWithContext(ctx context.Context, capacity int, sampleRate uint, sampleFormat sdr.SampleFormat) (*Pipe, error) {
	ctx, cancel := context.WithCancel(ctx)

	pipeReader, pipeWriter := sdr.PipeWithContext(ctx, sampleRate, sampleFormat)

	buf := make(chan sdr.Samples, capacity)
	pipe := &Pipe{
		ctx:    ctx,
		cancel: cancel,
		err:    nil,
		buf:    buf,

		sampleRate:   sampleRate,
		sampleFormat: sampleFormat,

		pipeReader: pipeReader,
		pipeWriter: pipeWriter,
	}
	go pipe.do()
	return pipe, nil
}

// vim: foldmethod=marker

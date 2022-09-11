// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2022
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
	"sync"

	"hz.tools/sdr"
)

// BufPipe2 is a new (more stable?) and experimental implementation
// of a buffered sdr.Pipe
type BufPipe2 struct {
	lock *sync.Mutex

	sampleRate   uint
	sampleFormat sdr.SampleFormat

	err    error
	closed bool
	buf    chan sdr.Samples

	pipeReader sdr.PipeReader
	pipeWriter sdr.PipeWriter
}

// SampleFormat implements the sdr.ReadWriter interface.
func (p *BufPipe2) SampleFormat() sdr.SampleFormat {
	return p.sampleFormat
}

// SampleRate implements the sdr.ReadWriter interface.
func (p *BufPipe2) SampleRate() uint {
	return p.sampleRate
}

// CloseWithError will close the pipe, and return the provided error on any
// subsequent call.
func (p *BufPipe2) CloseWithError(err error) error {
	p.err = err
	p.Close()
	return nil
}

// Close will close the BufPipe, which will remain readable until the end
// of buffered data.
func (p *BufPipe2) Close() error {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.closed {
		return nil
	}
	p.closed = true
	close(p.buf)
	return nil
}

// Read implements the sdr.ReadWriter interface.
func (p *BufPipe2) Read(s sdr.Samples) (int, error) {
	return p.pipeReader.Read(s)
}

// Write implements the sdr.ReadWriter interface.
func (p *BufPipe2) Write(s sdr.Samples) (int, error) {
	s2, i, err := dupe(s)
	if err != nil {
		return 0, err
	}

	p.lock.Lock()
	defer p.lock.Unlock()
	if p.closed {
		if p.err == nil {
			return 0, sdr.ErrPipeClosed
		}
		return 0, p.err
	}
	// TODO(paultag): add in blocking handling
	p.buf <- s2
	return i, nil
}

func (p *BufPipe2) do() {
	defer p.Close()
	defer func() { p.pipeWriter.CloseWithError(p.err) }()

	for {
		select {
		case s1, ok := <-p.buf:
			if !ok {
				p.lock.Lock()
				p.err = sdr.ErrPipeClosed
				p.lock.Unlock()
				return
			}
			_, err := p.pipeWriter.Write(s1)
			if err != nil {
				// If we caught an error and we don't have an error
				// ourselves (like a context error), we're going to
				// go ahead and set the error condition and bail.
				p.lock.Lock()
				if p.err == nil {
					p.err = err
					p.CloseWithError(err)
				}
				p.lock.Unlock()
				return
			}
		}
	}
}

// NewBufPipe2 will create a new stream.BufPipe, which wraps a normal sdr.Pipe,
// but writes will not block.
func NewBufPipe2(capacity int, sampleRate uint, sampleFormat sdr.SampleFormat) (*BufPipe2, error) {
	pipeReader, pipeWriter := sdr.Pipe(sampleRate, sampleFormat)
	buf := make(chan sdr.Samples, capacity)
	pipe := &BufPipe2{
		lock: &sync.Mutex{},
		err:  nil,
		buf:  buf,

		sampleRate:   sampleRate,
		sampleFormat: sampleFormat,

		pipeReader: pipeReader,
		pipeWriter: pipeWriter,
	}
	go pipe.do()
	return pipe, nil
}

// vim: foldmethod=marker

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
	"fmt"
	"sync"

	"hz.tools/sdr"
)

type standbyReader struct {
	lock *sync.Mutex

	// r may or may not be nil depending on if we're rx'ing or
	// not.
	r sdr.ReadCloser

	rx           sdr.Receiver
	sampleRate   uint
	sampleFormat sdr.SampleFormat
}

func (sr *standbyReader) SampleRate() uint {
	return sr.sampleRate
}

func (sr *standbyReader) SampleFormat() sdr.SampleFormat {
	return sr.sampleFormat
}

func (sr *standbyReader) Close() error {
	sr.lock.Lock()
	defer sr.lock.Unlock()
	if sr.r == nil {
		return nil
	}
	err := sr.r.Close()
	sr.r = nil
	return err
}

func (sr *standbyReader) Read(s sdr.Samples) (int, error) {
	sr.lock.Lock()
	defer sr.lock.Unlock()
	if sr.r == nil {
		r, err := sr.rx.StartRx()
		if err != nil {
			return 0, err
		}
		// TODO(paultag): Check SampleFormat against the Receiver here
		// too? Too much?
		if sr.sampleRate != r.SampleRate() {
			return 0, fmt.Errorf("StandbyReader.Read: SampleRate mismatch")
		}
		sr.r = r
	}
	return sr.r.Read(s)
}

// StandbyReader is a reusable ReadCloser which wraps an sdr.Receiver
// When IQ data is read from the Reader, it will StartRx, and read samples
// to the new underlying ReadCloser from StartRx. When "Close" is called,
// the underlying ReadCloser will be closed, but the StandbyReader remains
// usable.
//
// This enables easier management of the Receiver; less work needs to go
// into management of the state of the Receiver. Beware, this doesn't mean
// *no* management, just easier management.
//
// The underlying sdr.ReadCloser's SampleFormat is the SampleFormat returned
// by the Receiver, and the SampleRate is the SampleRate as read at the
// time on construction - and must not change.
func StandbyReader(rx sdr.Receiver) (sdr.ReadCloser, error) {
	sr, err := rx.GetSampleRate()
	if err != nil {
		return nil, err
	}

	return &standbyReader{
		lock:         &sync.Mutex{},
		rx:           rx,
		sampleRate:   sr,
		sampleFormat: rx.SampleFormat(),
	}, nil
}

// vim: foldmethod=marker

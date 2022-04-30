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

type standbyWriter struct {
	lock *sync.Mutex

	// w may or may not be nil depending on if we're tx'ing or
	// not.
	w sdr.WriteCloser

	tx           sdr.Transmitter
	sampleRate   uint
	sampleFormat sdr.SampleFormat
}

func (sw *standbyWriter) SampleRate() uint {
	return sw.sampleRate
}

func (sw *standbyWriter) SampleFormat() sdr.SampleFormat {
	return sw.sampleFormat
}

func (sw *standbyWriter) Close() error {
	sw.lock.Lock()
	defer sw.lock.Unlock()
	if sw.w == nil {
		return nil
	}
	err := sw.w.Close()
	sw.w = nil
	return err
}

func (sw *standbyWriter) Write(s sdr.Samples) (int, error) {
	sw.lock.Lock()
	defer sw.lock.Unlock()
	if sw.w == nil {
		w, err := sw.tx.StartTx()
		if err != nil {
			return 0, err
		}
		// TODO(paultag): Check SampleFormat against the Transmitter here
		// too? Too much?
		if sw.sampleRate != w.SampleRate() {
			return 0, fmt.Errorf("StandbyWriter.Write: SampleRate mismatch")
		}
		sw.w = w
	}
	return sw.w.Write(s)
}

// StandbyWriter is a reusable WriteCloser which wraps an sdr.Transmitter.
// When IQ data is written to the Writer, it will StartTx, and write samples
// to the new underlying WriteCloser from StartTx. When "Close" is called,
// the underlying WriteCloser will be closed, but the StandbyWriter remains
// usable.
//
// This enables easier management of the Transmitter; less work needs to go
// into management of the state of the Transmitter. Beware, this doesn't mean
// *no* management, just easier management.
//
// The underlying sdr.WriteCloser's SampleFormat is the SampleFormat returned
// by the Transmitter, and the SampleRate is the SampleRate as read at the
// time on construction - and must not change.
func StandbyWriter(tx sdr.Transmitter) (sdr.WriteCloser, error) {
	sr, err := tx.GetSampleRate()
	if err != nil {
		return nil, err
	}

	return &standbyWriter{
		lock:         &sync.Mutex{},
		tx:           tx,
		sampleRate:   sr,
		sampleFormat: tx.SampleFormat(),
	}, nil
}

// vim: foldmethod=marker

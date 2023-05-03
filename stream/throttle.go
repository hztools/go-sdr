// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2021
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
	"time"

	"hz.tools/sdr"
)

// Throttle will read the sdr.Reader's SampleRate, and throttle
// the stream to play the Reader back "real time".
func Throttle(r sdr.Reader) (sdr.Reader, error) {
	return ThrottleSecondsPerSecond(r, time.Second)
}

// ThrottleSecondsPerSecond will read the sdr.Reader's SampleRate, and throttle
// the stream to play the Reader back where the duration 'd' passes every
// second.
func ThrottleSecondsPerSecond(r sdr.Reader, d time.Duration) (sdr.Reader, error) {
	// Let's search for an easy window to size to use.
	var windowRate uint = 20
	for ; r.SampleRate()%windowRate == 0; windowRate++ {
	}
	pipeReader, pipeWriter := sdr.Pipe(r.SampleRate(), r.SampleFormat())
	buf, err := sdr.MakeSamples(r.SampleFormat(), int(r.SampleRate()/windowRate))
	if err != nil {
		return nil, err
	}
	go func() {
		defer pipeWriter.Close()
		clock := time.NewTicker(d / time.Duration(windowRate))
		defer clock.Stop()
		for {
			_, err := sdr.ReadFull(r, buf)
			if err != nil {
				return
			}
			select {
			case <-clock.C:
				pipeWriter.Write(buf)
			}
		}
	}()
	return pipeReader, nil
}

// vim: foldmethod=marker

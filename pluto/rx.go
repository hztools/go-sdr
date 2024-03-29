// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2020-2021
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

package pluto

import (
	"time"
	"unsafe"

	"hz.tools/sdr"
	"hz.tools/sdr/pluto/iio"
	"hz.tools/sdr/stream"
)

type rx struct {
	rxi        *iio.Channel
	rxq        *iio.Channel
	adc        *iio.Device
	windowSize int
}

func openRx(ictx *iio.Context, windowSize int) (*rx, error) {
	lpc, err := ictx.FindDevice(plutoRxName)
	if err != nil {
		return nil, err
	}

	rxi, err := lpc.FindChannel("voltage0", iio.ChannelDirectionRead)
	if err != nil {
		return nil, err
	}

	rxq, err := lpc.FindChannel("voltage1", iio.ChannelDirectionRead)
	if err != nil {
		return nil, err
	}

	return &rx{
		rxi:        rxi,
		rxq:        rxq,
		adc:        lpc,
		windowSize: windowSize,
	}, nil
}

type readCloser struct {
	writer        sdr.PipeWriter
	reader        sdr.PipeReader
	checkOverruns bool
	sdr           *Sdr
	buf           sdr.SamplesI16
}

func (rc *readCloser) Read(iq sdr.Samples) (int, error) {
	return rc.reader.Read(iq)
}

func (rc *readCloser) Close() error {
	rc.writer.Close()
	return nil
}

func (rc *readCloser) SampleRate() uint {
	return rc.reader.SampleRate()
}

func (rc *readCloser) SampleFormat() sdr.SampleFormat {
	return rc.reader.SampleFormat()
}

func (rc *readCloser) run() error {
	rx := rc.sdr.rx
	rx.rxi.Enable()
	rx.rxq.Enable()
	defer rx.rxi.Disable()
	defer rx.rxq.Disable()

	if rc.sdr.rxKernelBuffersCount != 0 {
		if err := rx.adc.SetKernelBuffersCount(rc.sdr.rxKernelBuffersCount); err != nil {
			return err
		}
	}

	ibuf, err := rx.adc.CreateBuffer(rc.buf.Length())
	if err != nil {
		return err
	}
	defer ibuf.Close()

	var (
		buf      = rc.buf
		nsamples int64

		// If we take more than 2 seconds, abort everything.
		deadline = time.Now().Add(time.Second * 2)
	)

	if err := rx.adc.ClearCheckBuffer(); err != nil {
		return err
	}

	for {
		i, err := ibuf.Refill()
		if err != nil {
			return err
		}
		buf := buf[:i/4]

		if rc.checkOverruns {
			if err := rx.adc.CheckBufferOverflow(); err != nil {
				if time.Now().After(deadline) {
					return iio.ErrOverrun
				}
				// Let's keep clearing the buffer until we can catch up with
				// ourselves. This won't go past the Check/Clear until it actually
				// gets a window without it having been dropped.
				if err == iio.ErrOverrun && nsamples == 0 {
					continue
				}
				return err
			}
		}

		i, err = ibuf.CopyToUnsafeFromBuffer(*rx.rxi, unsafe.Pointer(&buf[0]), buf.Size())
		if err != nil {
			return err
		}
		buf = buf[:i/4]
		buf.ShiftLSBToMSBBits(12)

		n, err := rc.writer.Write(buf)
		if err != nil {
			return err
		}
		nsamples += int64(n)
	}
}

// StartRx implements the sdr.Sdr interface.
func (s *Sdr) StartRx() (sdr.ReadCloser, error) {
	ring, err := stream.NewRingBuffer(
		s.samplesPerSecond,
		sdr.SampleFormatI16,
		stream.RingBufferOptions{
			Slots:      32,
			SlotLength: s.rxWindowSize,
			BlockReads: true,
		},
	)
	if err != nil {
		return nil, err
	}

	rc := &readCloser{
		checkOverruns: s.checkOverruns,
		writer:        ring,
		reader:        ring,
		sdr:           s,
		buf:           make(sdr.SamplesI16, s.rxWindowSize),
	}

	go func() {
		if err := rc.run(); err != nil {
			ring.CloseWithError(err)
		}
	}()

	return rc, nil

}

// vim: foldmethod=marker

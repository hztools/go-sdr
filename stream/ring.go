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

var (
	// // ErrRingBufferOverrun will be returned if a Write operation on a
	// // Ring Buffer catches up to the Read head.
	// //
	// // This error is only returned if BlockWrites is set to False.
	// ErrRingBufferOverrun error = fmt.Errorf("RingBuffer: Buffer Overrun")

	// ErrRingBufferUnderrun will be returned if a Read operation on a
	// Ring Buffer catches up to the Write head. This can be a temporary
	// state, and subsequent Read operations will be error free if a
	// Slot is written to.
	//
	// This error is only returned if BlockReads is set to False.
	ErrRingBufferUnderrun error = fmt.Errorf("RingBuffer: Buffer Underrun")
)

// RingBufferOptions contains configurable options for a Ring Buffer.
type RingBufferOptions struct {
	// Slots are the nuber of IQ slots in the Ring Buffer.
	Slots int

	// SlotLength is the max number of IQ samples the Ring Buffer can
	// store, per slot.
	SlotLength int

	// BlockReads will force a wait (rather than an ErrRingBufferUnderrun)
	// if the Read cursor has caught up with the Write cursor.
	BlockReads bool
}

// RingBuffer is an IQ Ring Buffer, where no allocations have to happen to write,
// designed to take high frequency input data without dealing with things like
// channel latency or goroutine scheduling.
type RingBuffer struct {
	cond *sync.Cond
	lock *sync.Mutex

	buf  sdr.Samples
	bufn []int

	closed bool
	err    error

	slots   int
	slotLen int

	ridx int
	widx int

	format sdr.SampleFormat
	rate   uint

	// stats
	overruns  int
	underruns int

	opts RingBufferOptions
}

// slot will return the nth slot
func (rb *RingBuffer) slot(n int) (sdr.Samples, error) {
	if n >= rb.slots {
		return nil, fmt.Errorf("RingBuffer: Slot is out of range")
	}
	base := (n * rb.slotLen)
	return rb.buf.Slice(base, base+rb.slotLen), nil
}

// Read implements the sdr.Reader interface.
func (rb *RingBuffer) Read(buf sdr.Samples) (int, error) {
	if buf.Length() < rb.slotLen {
		return 0, fmt.Errorf("RingBuffer: Slot is larger than the target Read buffer")
	}

	rb.lock.Lock()
	defer rb.lock.Unlock()

	if rb.ridx == rb.widx {
		// This is reached when there's nothing in the buffer.

		// Initial closed check; if we're closed, let's bail here.
		if rb.closed {
			// if we're closed, dump the error state.
			if rb.err == nil {
				return 0, sdr.ErrPipeClosed
			}
			return 0, rb.err
		}

		if !rb.opts.BlockReads {
			rb.underruns++
			return 0, ErrRingBufferUnderrun
		}

		for rb.ridx == rb.widx {
			// Attempt to aquire the lock until we have data that we
			// can read from the next slot.
			rb.cond.Wait()
		}

		// Check if we were closed since we lost the lock.
		if rb.closed {
			// if we're closed, dump the error state.
			if rb.err == nil {
				return 0, sdr.ErrPipeClosed
			}
			return 0, rb.err
		}
	}

	idx := rb.ridx
	rb.ridx = (rb.ridx + 1) % rb.slots
	slot, err := rb.slot(idx)
	if err != nil {
		return 0, err
	}
	slot = slot.Slice(0, rb.bufn[idx])
	n, err := sdr.CopySamples(buf, slot)

	// zero out the buffer length after read
	rb.bufn[idx] = 0

	return n, err
}

// Write implements the sdr.Writer interface.
func (rb *RingBuffer) Write(buf sdr.Samples) (int, error) {
	if buf.Length() > rb.slotLen {
		return 0, fmt.Errorf("RingBuffer: Slot is larger than provided Write buffer")
	}

	rb.lock.Lock()
	defer rb.lock.Unlock()

	// sorry, no :)
	if rb.closed {
		if rb.err == nil {
			return 0, sdr.ErrPipeClosed
		}
		return 0, rb.err
	}

	nwidx := (rb.widx + 1) % rb.slots
	if nwidx == rb.ridx {
		rb.overruns++

		// TODO: add in ErrRingBufferOverrun toggles.

		// to overwrite, we need to drop the read slot by
		// dropping the oldest slot, bump forward by one.
		rb.ridx = (rb.ridx + 1) % rb.slots
	}

	idx := rb.widx
	rb.widx = nwidx // advance the write index

	slot, err := rb.slot(idx)
	if err != nil {
		return 0, err
	}
	rb.bufn[idx] = buf.Length()
	n, err := sdr.CopySamples(slot, buf)
	if !rb.opts.BlockReads {
		rb.cond.Signal()
	}
	return n, err
}

// StatsOverrun will return the count of Overruns that have taken place.
func (rb *RingBuffer) StatsOverrun() int {
	return rb.overruns
}

// StatsUnderrun will return the count of Underruns that have taken place.
func (rb *RingBuffer) StatsUnderrun() int {
	return rb.underruns
}

// CloseWithError will set the error state on the Ring Buffer.
func (rb *RingBuffer) CloseWithError(err error) error {
	rb.err = err
	return rb.Close()
}

// Close implements the sdr.Closer interface.
func (rb *RingBuffer) Close() error {
	rb.closed = true
	rb.cond.Signal()
	return nil
}

// SampleFormat implements the sdr.ReadWriteCloser interface.
func (rb *RingBuffer) SampleFormat() sdr.SampleFormat {
	return rb.format
}

// SampleRate implements the sdr.ReadWriteCloser interface.
func (rb *RingBuffer) SampleRate() uint {
	return rb.rate
}

// NewRingBuffer will create a RingBuffer with the provided options
func NewRingBuffer(
	rate uint,
	format sdr.SampleFormat,
	opts RingBufferOptions,
) (*RingBuffer, error) {
	var (
		lock = &sync.Mutex{}
		cond = sync.NewCond(lock)
	)

	buf, err := sdr.MakeSamples(format, opts.Slots*opts.SlotLength)
	if err != nil {
		return nil, err
	}

	return &RingBuffer{
		cond: cond,
		lock: lock,
		buf:  buf,
		bufn: make([]int, opts.Slots),

		slots:   opts.Slots,
		slotLen: opts.SlotLength,

		rate:   rate,
		format: format,
		opts:   opts,
	}, nil
}

// vim: foldmethod=marker

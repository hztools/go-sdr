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
	"unsafe"

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

	// IQBufferAllocator will be passed the configured RingBufferOptions,
	// and allocate an sdr.Samples object that is long enough for the
	// Buffer (Slots*SlotLength at minimum). If Nil, this will use
	// the stock Allocator.
	IQBufferAllocator func(sdr.SampleFormat, RingBufferOptions) (sdr.Samples, error)

	// IQBufferSlotSlicer is responsible for returning a slice of the IQ Buffer
	// allocated to a specific slot. If Nil, this will use the stock Slicer.
	IQBufferSlotSlicer func(sdr.Samples, int, RingBufferOptions) sdr.Samples
}

func (opts RingBufferOptions) getIQBufferSlotSlicer() func(sdr.Samples, int, RingBufferOptions) sdr.Samples {
	if opts.IQBufferSlotSlicer != nil {
		return opts.IQBufferSlotSlicer
	}
	return func(s sdr.Samples, id int, opts RingBufferOptions) sdr.Samples {
		base := (id * opts.slotLength())
		return s.Slice(base, base+opts.slotLength())
	}
}

// getIQBufferAllocator will return the RingBuffer IQ Allocator configured
// in the RingBufferOptions; or return the default IQ Allocator that uses
// sdr.MakeSamples.
func (opts RingBufferOptions) getIQBufferAllocator() func(sdr.SampleFormat, RingBufferOptions) (sdr.Samples, error) {
	if opts.IQBufferAllocator != nil {
		return opts.IQBufferAllocator
	}
	return func(format sdr.SampleFormat, opts RingBufferOptions) (sdr.Samples, error) {
		return sdr.MakeSamples(format, opts.Slots*opts.SlotLength)
	}
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

	ridx int
	widx int

	rate uint

	opts RingBufferOptions
}

// slots will return the configured Slot count.
func (opts RingBufferOptions) slots() int {
	return opts.Slots
}

func (rb *RingBuffer) slots() int {
	return rb.opts.slots()
}

// slotLength will return the configured Slot Length.
func (opts RingBufferOptions) slotLength() int {
	return opts.SlotLength
}

func (rb *RingBuffer) slotLength() int {
	return rb.opts.slotLength()
}

// slot will return the nth slot
func (rb *RingBuffer) slot(n int) (sdr.Samples, error) {
	if n >= rb.slots() {
		return nil, fmt.Errorf("RingBuffer: Slot is out of range")
	}
	return rb.opts.getIQBufferSlotSlicer()(rb.buf, n, rb.opts), nil
}

// advanceReadCursor (UNSAFE) will return the slot number of the next slot to be read.
//
// callers MUST have the mutex.
//
// If -1 is returned, there is no unread data in the Ring Buffer, otherwise the
// ID is the next slot to be read. As a side-effect, this will advance the read
// cursor if a slot is returned.
func (rb *RingBuffer) advanceReadCursor() int {
	if rb.ridx == rb.widx {
		return -1
	}
	idx := rb.ridx
	rb.ridx = (rb.ridx + 1) % rb.slots()
	return idx
}

// advanceWriteCursor (UNSAFE) will return the slot number of the next slot to
// be written to.
//
// callers MUST have the mutex.
//
// if the 'overwrite' boolean is true, this may result in the read cursor being
// advanced, and a read slot being dropped to make room. if the overwrite bool
// is false, -1 will be returned.
//
// the 0th argument returned is the write slot. If the argument is -1, the queue
// is full and we can not overwrite.
//
// the 1st argument returned is a boolean indicating if an overrun has happened,
// resulting in a read drop
func (rb *RingBuffer) advanceWriteCursor(overwrite bool) (int, bool) {
	nwidx := (rb.widx + 1) % rb.slots()
	if nwidx == rb.ridx {
		// right, so we're full. let's consult the overwrite boolean.
		if overwrite {
			rb.advanceReadCursor()
			id, _ := rb.advanceWriteCursor(overwrite)
			return id, true
		}
		// if we can't overwrite, lets give up. no slot, and we didn't drop
		// any data.
		return -1, false
	}
	idx := rb.widx
	rb.widx = nwidx // advance the write index
	return idx, false
}

func (rb *RingBuffer) getErr() error {
	if !rb.closed {
		return nil
	}

	if rb.err == nil {
		return sdr.ErrPipeClosed
	}
	return rb.err
}

// Read implements the sdr.Reader interface.
func (rb *RingBuffer) Read(buf sdr.Samples) (int, error) {
	if buf.Length() < rb.slotLength() {
		return 0, fmt.Errorf("RingBuffer: Slot is larger than the target Read buffer")
	}

	rb.lock.Lock()
	defer rb.lock.Unlock()

	id := rb.advanceReadCursor()
	if id == -1 {
		// This is reached when there's nothing in the buffer.

		if err := rb.getErr(); err != nil {
			return 0, err
		}

		// If we don't block reads, let's immediately return an underrun.
		if !rb.opts.BlockReads {
			return 0, ErrRingBufferUnderrun
		}

		// attempt to move the cursor forward.
		for ; id == -1; id = rb.advanceReadCursor() {
			// Attempt to aquire the lock until we have data that we
			// can read from the next slot.
			rb.cond.Wait()
		}

		if err := rb.getErr(); err != nil {
			return 0, err
		}
	}

	slot, err := rb.slot(id)
	if err != nil {
		return 0, err
	}
	slot = slot.Slice(0, rb.bufn[id])
	n, err := sdr.CopySamples(buf, slot)

	// zero out the buffer length after read
	rb.bufn[id] = 0

	return n, err
}

// Write implements the sdr.Writer interface.
func (rb *RingBuffer) Write(buf sdr.Samples) (int, error) {
	if buf.Length() > rb.slotLength() {
		return 0, fmt.Errorf("RingBuffer: Slot is larger than provided Write buffer")
	}

	rb.lock.Lock()
	defer rb.lock.Unlock()

	// sorry, no :)
	if err := rb.getErr(); err != nil {
		return 0, err
	}

	// advance the write header, blow away any existing data for now
	// TODO: add in a block toggle.
	id, _ := rb.advanceWriteCursor(true)

	slot, err := rb.slot(id)
	if err != nil {
		return 0, err
	}
	n, err := sdr.CopySamples(slot, buf)
	rb.bufn[id] = n
	if !rb.opts.BlockReads {
		rb.cond.Signal()
	}
	return n, err
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
	return rb.buf.Format()
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

	if opts.Slots == 0 || opts.SlotLength == 0 {
		return nil, fmt.Errorf("stream.NewRingBuffer: Slots and SlotLength must be set to a value other than 0")
	}

	buf, err := opts.getIQBufferAllocator()(format, opts)
	if err != nil {
		return nil, err
	}

	if buf.Length() < opts.Slots*opts.SlotLength {
		return nil, fmt.Errorf("stream.NewRingBuffer: opts.IQBufferAllocator did not return enough IQ space.")
	}

	return &RingBuffer{
		cond: cond,
		lock: lock,
		buf:  buf,
		bufn: make([]int, opts.Slots),
		rate: rate,
		opts: opts,
	}, nil
}

// Unsafe wrapper here for I/O perf critical zero-copy applications.
//
// Here be dragons.

// UnsafeRingBuffer is a wrapper around a RingBuffer that allows a few specific
// and very unsafe things for the sake of I/O critical direct access given a-priori
// knowledge of the IQ Buffer and a deep understanding of the thread safty.
type UnsafeRingBuffer struct {
	*RingBuffer
}

// WritePeek will return the underlying Slot the Write cursor will write to
// next, but without advancing the Write cursor. By the time this function
// returns, the index can be wrong if any other Writes are happening to this
// Ring Buffer.
func (urb *UnsafeRingBuffer) WritePeek() int {
	return urb.widx
}

// WritePeekUnsafePointer will return an unsafe.Pointer pointing to
// the 0th element of the Slot. This method has the same caviats as WritePeek,
// and more, since it's wildly unsafe.
func (urb *UnsafeRingBuffer) WritePeekUnsafePointer() unsafe.Pointer {
	b := sdr.MustUnsafeSamplesAsBytes(
		urb.opts.getIQBufferSlotSlicer()(urb.buf, urb.widx, urb.opts),
	)
	return unsafe.Pointer(&b[0])
}

// WritePoke will write the next slot (blindly) assuming that the caller
// has used WritePeek to figure out what cell we'll be using next, infer the
// slice based on the layout, written data there, and ensured that no race
// conditions can cause any other Writes to the RingBuffer.
func (urb *UnsafeRingBuffer) WritePoke(n int) {
	urb.lock.Lock()
	defer urb.lock.Unlock()

	id, _ := urb.advanceWriteCursor(true)
	urb.bufn[id] = n
	if !urb.opts.BlockReads {
		urb.cond.Signal()
	}
}

// UnsafeRingBuffer will return the underlying ring buffer as created by
// the IQBufferAllocator.
func (urb *UnsafeRingBuffer) UnsafeGetIQBuffer() sdr.Samples {
	return urb.buf
}

// NewUnsafeRingBuffer will create a RingBuffer wrapper around the provided
// RingBuffer, but will enable some very shady helpers to mutate the RingBuffer
// given a deep understanding of the underlying magic.
func NewUnsafeRingBuffer(rb *RingBuffer) *UnsafeRingBuffer {
	return &UnsafeRingBuffer{rb}
}

// vim: foldmethod=marker

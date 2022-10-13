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

package uhd

// This file contains some -- frankly -- vile hacks. This will take control
// of a stream.RingBuffer, replace the IQ Allocation with C controlled memory,
// and use the Unsafe wrapper to allow the C UHD API to write directly into
// the RingBuffer without taking the Read mutex, or copying IQ samples each
// loop.
//
// If *ANYTHING* Writes to the RingBuffer behind our back, it *will* corrupt
// it.

// #include <malloc.h>
import "C"

import (
	"hz.tools/sdr"
	"hz.tools/sdr/internal/yikes"
	"hz.tools/sdr/stream"
)

// slicer will do the IQ Buffer slicing for us.
func slicer(s sdr.Samples, id int, opts stream.RingBufferOptions) sdr.Samples {
	base := id * opts.SlotLength
	return s.Slice(base, base+opts.SlotLength)
}

// slicerBytes will return the byte index to get to the right Slot's 0th IQ sample.
func slicerBytes(sf sdr.SampleFormat, id int, opts stream.RingBufferOptions) int {
	return sf.Size() * id * opts.SlotLength
}

// newCUnsafeRingBuffer will create a new RingBuffer, but will override
// the IQBuffer Allocator and Slicer to use C memory
func newCUnsafeRingBuffer(rate uint, format sdr.SampleFormat, opts stream.RingBufferOptions) (*stream.UnsafeRingBuffer, error) {
	opts.IQBufferAllocator = func(format sdr.SampleFormat, opts stream.RingBufferOptions) (sdr.Samples, error) {
		iqLength := opts.Slots * opts.SlotLength
		iqSize := iqLength * format.Size()
		buf := C.malloc(C.size_t(iqSize))
		iq, err := yikes.Samples(uintptr(buf), iqLength, format)
		if err != nil {
			return nil, err
		}
		return iq, nil
	}
	rb, err := stream.NewRingBuffer(rate, format, opts)
	if err != nil {
		return nil, err
	}
	return stream.NewUnsafeRingBuffer(rb), nil
}

// vim: foldmethod=marker

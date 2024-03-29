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

package iio

// #cgo pkg-config: libiio
//
// #include <iio.h>
import "C"

import (
	"fmt"
	"syscall"
	"unsafe"

	"hz.tools/sdr/yikes"
)

// Buffer wraps an iio_buffer, which allows for reading samples from and writing
// data to the underlying device.
type Buffer struct {
	closed *bool
	handle *C.struct_iio_buffer
}

// Close will destroy the handle to the iio_buffer.
func (b Buffer) Close() error {
	if *b.closed {
		return nil
	}
	C.iio_buffer_destroy(b.handle)
	*b.closed = true
	return nil
}

// Step will return the iio_buffer_step, which is the size of each coherent
// sample.
func (b Buffer) Step() uintptr {
	return uintptr(C.iio_buffer_step(b.handle))
}

// CopyToBufferFromUnsafe will copy data in a very unsafe and deeply bad way to a given
// pointer and size in bytes.
func (b Buffer) CopyToBufferFromUnsafe(chn Channel, ptr unsafe.Pointer, size int) (int, error) {
	base := uintptr(C.iio_buffer_first(b.handle, chn.handle))
	end := uintptr(C.iio_buffer_end(b.handle))
	totalBytes := int(end - base)
	if totalBytes < 0 {
		return 0, fmt.Errorf("iio: internal error during Buffer.CopyToUnsafe")
	}

	if totalBytes == 0 {
		return 0, nil
	}

	bufMemory := yikes.GoBytes(base, totalBytes)
	targetMemory := yikes.GoBytes(uintptr(ptr), size)

	i := copy(bufMemory, targetMemory)
	return i, nil
}

// CopyToUnsafeFromBuffer will copy data in a very unsafe and deeply bad way to a given
// pointer and size in bytes.
func (b Buffer) CopyToUnsafeFromBuffer(chn Channel, ptr unsafe.Pointer, size int) (int, error) {
	base := uintptr(C.iio_buffer_first(b.handle, chn.handle))
	end := uintptr(C.iio_buffer_end(b.handle))
	totalBytes := int(end - base)
	if totalBytes < 0 {
		return 0, fmt.Errorf("iio: internal error during Buffer.CopyToUnsafe")
	}

	if totalBytes == 0 {
		return 0, nil
	}

	bufMemory := yikes.GoBytes(base, totalBytes)
	targetMemory := yikes.GoBytes(uintptr(ptr), size)

	i := copy(targetMemory, bufMemory)
	return i, nil
}

// PushPartial will push the data written to the Buffer (from start to end) to the
// Device.
func (b Buffer) PushPartial(length int) (int, error) {
	i := C.iio_buffer_push_partial(b.handle, C.size_t(length))
	if i < 0 {
		return 0, syscall.Errno(-i)
	}
	return int(i), nil
}

// Push will push the data written to the Buffer (from start to end) to the
// Device.
func (b Buffer) Push() (int, error) {
	i := C.iio_buffer_push(b.handle)
	if i < 0 {
		return 0, syscall.Errno(-i)
	}
	return int(i), nil
}

// Refill will fill the Buffer up with samples from the backing device.
func (b Buffer) Refill() (int, error) {
	i := C.iio_buffer_refill(b.handle)
	if i < 0 {
		return 0, syscall.Errno(-i)
	}
	return int(i), nil
}

func (d Device) createBuffer(samplesCount int, cyclic bool) (*Buffer, error) {
	buf, err := C.iio_device_create_buffer(
		d.handle,
		C.size_t(samplesCount),
		C.bool(cyclic),
	)
	if buf == nil {
		return nil, err
	}
	var closed bool
	return &Buffer{
		handle: buf,
		closed: &closed,
	}, nil
}

// CreateBuffer will create an iio_buffer from a given device.
//
// At least one channel must be Enabled prior to this call.
func (d Device) CreateBuffer(samplesCount int) (*Buffer, error) {
	return d.createBuffer(samplesCount, false)
}

// CreateCyclicBuffer will create an iio_buffer from a given device.
//
// At least one channel must be Enabled prior to this call.
func (d Device) CreateCyclicBuffer(samplesCount int) (*Buffer, error) {
	return d.createBuffer(samplesCount, true)
}

// vim: foldmethod=marker

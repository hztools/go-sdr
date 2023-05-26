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
// #include <stdlib.h>
import "C"

import (
	"fmt"
	"syscall"
	"unsafe"
)

// FindDevice will find a device with a given name
func (c Context) FindDevice(name string) (*Device, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	dev := C.iio_context_find_device(c.handle, cName)
	if dev == nil {
		return nil, fmt.Errorf("iio: device '%s' is not known", name)
	}

	return &Device{name: name, handle: dev}, nil
}

// Device represents an iio_device within an iio_context.
type Device struct {
	name   string
	handle *C.struct_iio_device
}

// String will return the name of the Device.
func (d Device) String() string {
	return d.name
}

var (
	// ErrOverrun will be returned if samples have been dropped on the
	// receive path.
	ErrOverrun = fmt.Errorf("iio: iq overrun")

	// ErrUnderrun will be returned if the buffer ran out of samples while
	// transmitting.
	ErrUnderrun = fmt.Errorf("iio: iq underrun")
)

// SetKernelBuffersCount will configure the number of kernelspace buffers
// to be used when transfering data to and from the device. The default is 4.
func (d Device) SetKernelBuffersCount(nbuf uint) error {
	errno := C.iio_device_set_kernel_buffers_count(d.handle, C.uint(nbuf))
	if errno == 0 {
		return nil
	}
	return syscall.Errno(-errno)
}

// ClearCheckBuffer will clear registry flags.
func (d Device) ClearCheckBuffer() error {
	errno := C.iio_device_reg_write(d.handle, 0x80000088, 0x06)
	if errno == 0 {
		return nil
	}
	return syscall.Errno(-errno)
}

// CheckBufferOverflow will check to see if there was an overrun when streaming
// IQ samples.
//
// If there was an overflow condition, this will return an iio.ErrOverrun. If
// there was an error fetching the overflow stauts, this will return an errno.
func (d Device) CheckBufferOverflow() error {
	var val C.uint32_t
	errno := C.iio_device_reg_read(d.handle, 0x80000088, &val)
	if errno != 0 {
		return syscall.Errno(-errno)
	}
	if (val & 4) == 4 {
		C.iio_device_reg_write(d.handle, 0x80000088, 4)
		return ErrOverrun
	}
	return nil
}

// CheckBufferUnderflow will check to see if there was an underrun when
// streaming IQ samples.
//
// If there was an underflow condition, this will return an iio.ErrUnderrun. If
// there was an error fetching the overflow stauts, this will return an errno.
func (d Device) CheckBufferUnderflow() error {
	var val C.uint32_t
	errno := C.iio_device_reg_read(d.handle, 0x80000088, &val)
	if errno != 0 {
		return syscall.Errno(-errno)
	}
	if (val & 1) == 1 {
		C.iio_device_reg_write(d.handle, 0x80000088, 1)
		return ErrUnderrun
	}
	return nil
}

// WriteDebugInt64 will write a debug int64 channel attribute to the backing device.
func (d Device) WriteDebugInt64(name string, value int64) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	errno := C.iio_device_debug_attr_write_longlong(
		d.handle,
		cName,
		C.longlong(value),
	)
	if errno == 0 {
		return nil
	}
	return syscall.Errno(-errno)
}

// vim: foldmethod=marker

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

// Channel is a stream of information to be read from the device.
type Channel struct {
	name      string
	direction ChannelDirection
	handle    *C.struct_iio_channel
}

// Enable will enable this channel for read or write.
func (c Channel) Enable() error {
	C.iio_channel_enable(c.handle)
	return nil
}

// Disable will enable this channel for read or write.
func (c Channel) Disable() error {
	C.iio_channel_disable(c.handle)
	return nil
}

// String will return the Channel's name.
func (c Channel) String() string {
	return c.name
}

// ReadInt64 will write an int64 chanel attribute to the backing device.
func (c Channel) ReadInt64(name string) (int64, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	cValue := C.longlong(0)

	errno := C.iio_channel_attr_read_longlong(
		c.handle,
		cName,
		&cValue,
	)
	if errno != 0 {
		return 0, syscall.Errno(-errno)
	}

	return int64(cValue), nil
}

// WriteInt64 will write an int64 chanel attribute to the backing device.
func (c Channel) WriteInt64(name string, value int64) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	errno := C.iio_channel_attr_write_longlong(
		c.handle,
		cName,
		C.longlong(value),
	)
	if errno == 0 {
		return nil
	}
	return syscall.Errno(-errno)
}

// WriteFloat64 will write an float64 chanel attribute to the backing device.
// (this is otherwise known as WriteDouble, but I've chosen the Go types here)
func (c Channel) WriteFloat64(name string, value float64) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	errno := C.iio_channel_attr_write_double(
		c.handle,
		cName,
		C.double(value),
	)
	if errno == 0 {
		return nil
	}
	return syscall.Errno(-errno)
}

// WriteBool will write a boolean channel attribute to the backing device.
func (c Channel) WriteBool(name string, value bool) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	errno := C.iio_channel_attr_write_bool(c.handle, cName, C.bool(value))
	if errno < 0 {
		return syscall.Errno(-errno)
	}
	// otherwise it's the number of bytes, which we're not interested in
	// at this time
	return nil
}

// WriteString will write a string channel attribute to the backing device.
func (c Channel) WriteString(name, value string) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	errno := C.iio_channel_attr_write(c.handle, cName, cValue)
	if errno < 0 {
		return syscall.Errno(-errno)
	}
	// otherwise it's the number of bytes, which we're not interested in
	// at this time
	return nil
}

// ChannelDirection is the direction of the channel, either able to "read"
// or "write" to and from the channel. This is slightly different than the
// underlying iio library, since this is called "output", but the quirk here
// is that in an RF capacity, the "output" channel is used to read / receive
// data, not send it.
type ChannelDirection bool

const (
	// ChannelDirectionWrite is a channel that we write into.
	ChannelDirectionWrite ChannelDirection = true

	// ChannelDirectionRead is a channel that we read from.
	ChannelDirectionRead ChannelDirection = false
)

// FindChannel will find a channel with the given name and direction.
func (d Device) FindChannel(name string, direction ChannelDirection) (*Channel, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var output C.bool = false
	switch direction {
	case ChannelDirectionWrite:
		output = true
	}

	chn := C.iio_device_find_channel(d.handle, cName, output)
	if chn == nil {
		return nil, fmt.Errorf("iio: channel '%s' is not known", name)
	}
	return &Channel{name: name, direction: direction, handle: chn}, nil
}

// vim: foldmethod=marker

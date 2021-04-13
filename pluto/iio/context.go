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
// #include <malloc.h>
import "C"

import (
	"unsafe"

	"hz.tools/sdr/debug"
)

// Context is a bit of hardware we're talking to.
//
// TODO(paultag): Rename this from Context, since that's overloaded in Go land.
// maybe platform?
type Context struct {
	name   string
	handle *C.struct_iio_context
	closed *bool
}

// String will return the name of the device.
func (c Context) String() string {
	return c.name
}

// Close will destroy all resources held by the wrapper (calling
// iio_context_destroy)
func (c Context) Close() error {
	if *c.closed {
		return nil
	}
	C.iio_context_destroy(c.handle)
	*c.closed = true
	debug.PprofCloser().Remove(c.handle)
	return nil
}

// Attr will fetch a specific attribute from the context.
func (c Context) Attr(name string) *string {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	cValue := C.iio_context_get_attr_value(
		c.handle,
		cName,
	)

	if cValue == nil {
		return nil
	}

	value := C.GoString(cValue)
	return &value
}

// Open will open the URI as a device.
func Open(uri string) (*Context, error) {
	cURI := C.CString(uri)
	defer C.free(unsafe.Pointer(cURI))

	ctx, err := C.iio_create_context_from_uri(cURI)
	if ctx == nil {
		return nil, err
	}
	debug.PprofCloser().Add(ctx, 1)

	var closed bool
	return &Context{
		name:   uri,
		handle: ctx,
		closed: &closed,
	}, nil
}

// vim: foldmethod=marker

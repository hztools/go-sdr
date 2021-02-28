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

package sdr

import (
	"fmt"
	"unsafe"
)

// UnsafeSamplesAsBytes is a very dangerous function.
//
// Use of the function should be *seriously* discouraged and not used to every
// extent possible. This is only to be used at carefully controlled boundaries
// where safe high-level primatives can't be used for a serious technical
// purpose.
func UnsafeSamplesAsBytes(buf Samples) ([]byte, error) {
	var base uintptr

	switch buf := buf.(type) {
	case SamplesI16:
		base = uintptr(unsafe.Pointer(&buf[0]))
	case SamplesC64:
		base = uintptr(unsafe.Pointer(&buf[0]))
	case SamplesU8:
		base = uintptr(unsafe.Pointer(&buf[0]))
	default:
		return nil, ErrSampleFormatUnknown
	}

	size := buf.Size()
	var b = struct {
		addr uintptr
		len  int
		cap  int
	}{base, size, size}
	return *(*[]byte)(unsafe.Pointer(&b)), nil
}

// MustUnsafeSamplesAsBytes will call the very dangerous UnsafeSamplesAsBytes
// function, and add even more unsafe and dangerous behavior on top -- namely,
// a panic if the Samples type can not be represented as a byte slice.
//
// Use of the function should be *seriously* discouraged and not used to every
// extent possible.
func MustUnsafeSamplesAsBytes(buf Samples) []byte {
	bufBytes, err := UnsafeSamplesAsBytes(buf)
	if err != nil {
		panic(fmt.Sprintf("sdr.MustUnsafeSamplesAsBytes: %s", err))
	}
	return bufBytes
}

// vim: foldmethod=marker

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

package yikes

import (
	"unsafe"

	"hz.tools/sdr"
)

// GoBytes works like C.GoBytes, but it allows for mutating the C byte array
// from Go. This is wildly unsafe, and something that needs to be very carefully
// applied to problems, but is generally going to be used at i/o boundaries,
// specifically on the tx paths.
func GoBytes(
	base uintptr,
	size int,
) []byte {
	var b = struct {
		base uintptr
		len  int
		cap  int
	}{base, size, size}
	return *(*[]byte)(unsafe.Pointer(&b))
}

// Samples will convert a pointer and a length into a Samples buffer of the
// provided underlying type.
//
// This is literally the worst. please, for the love of god, do not use this
// unless you absolutely have to.
func Samples(base uintptr, length int, sampleFormat sdr.SampleFormat) (sdr.Samples, error) {
	// Similar to above, we're going to allocate a slice header here,
	// and set it to the size of the underlying type.
	var b = struct {
		base uintptr
		len  int
		cap  int
	}{base, length, length}

	switch sampleFormat {
	case sdr.SampleFormatC64:
		return *(*sdr.SamplesC64)(unsafe.Pointer(&b)), nil
	case sdr.SampleFormatI16:
		return *(*sdr.SamplesI16)(unsafe.Pointer(&b)), nil
	case sdr.SampleFormatI8:
		return *(*sdr.SamplesI8)(unsafe.Pointer(&b)), nil
	case sdr.SampleFormatU8:
		return *(*sdr.SamplesU8)(unsafe.Pointer(&b)), nil
	default:
		return nil, sdr.ErrSampleFormatUnknown
	}
}

// vim: foldmethod=marker

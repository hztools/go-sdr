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
	"io"
)

// CopySamples is the interface version of `copy`, which is type-aware.
//
// This is used when you want to copy samples between two buffers of the same
// type. This can't be used for conversion.
func CopySamples(dst, src Samples) (int, error) {
	if dst.Format() != src.Format() {
		return 0, ErrSampleFormatMismatch
	}

	switch dst := dst.(type) {
	case SamplesU8:
		src := src.(SamplesU8)
		return copy(dst, src), nil
	case SamplesI8:
		src := src.(SamplesI8)
		return copy(dst, src), nil
	case SamplesI16:
		src := src.(SamplesI16)
		return copy(dst, src), nil
	case SamplesC64:
		src := src.(SamplesC64)
		return copy(dst, src), nil
	default:
		return 0, ErrSampleFormatUnknown
	}
}

// Copy will copy samples from the src sdr.Reader to the dst sdr.Writer.
//
// The Reader and Writer must be of the same SampleFormat. If not, that will
// return an error, and the caller should explicitly define how and where to
// convert the two formats.
func Copy(dst Writer, src Reader) (int64, error) {
	if dst.SampleFormat() != src.SampleFormat() {
		return 0, ErrSampleFormatMismatch
	}
	return copyBuffer(dst, src, nil)
}

// CopyBuffer will copy samples from the src sdr.Reader to the dst sdr.Writer
// using the provided Buffer.
func CopyBuffer(dst Writer, src Reader, buf Samples) (int64, error) {
	if dst.SampleFormat() != src.SampleFormat() {
		return 0, ErrSampleFormatMismatch
	}
	if dst.SampleFormat() != buf.Format() {
		return 0, ErrSampleFormatMismatch
	}
	return copyBuffer(dst, src, buf)
}

// copyBuffer will copy data from the src into the dst, using the buffer `buf`
// to move the data. If buf is nil, the size will be 1024*32.
func copyBuffer(dst Writer, src Reader, buf Samples) (int64, error) {
	var (
		err     error
		written int64
	)

	if buf == nil {
		size := 32 * 1024
		buf, err = MakeSamples(dst.SampleFormat(), size)
		if err != nil {
			return 0, err
		}
	}

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf.Slice(0, nr))
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}

// vim: foldmethod=marker

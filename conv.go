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
)

var (
	// ErrConversionNotImplemented will be returned if the Sample type is
	// unable to be converted into the desired target format.
	ErrConversionNotImplemented error = fmt.Errorf("sdr.Convert: unknown format conversion")
)

// ConvertBuffer the provided Samples to the desired output format.
//
// The conversion will happen in CPU, and this format can be a little slow,
// but it will get the job done, and beats the heck out of having to worry
// about the underlying data format.
//
// In the event that the desired format is the same as the provided format
// this function will copy the source samples to the target buffer.
func ConvertBuffer(dst, src Samples) error {
	if src.Format() == dst.Format() {
		_, err := CopySamples(dst, src)
		return err
	}

	if src.Length() > dst.Length() {
		return ErrDstTooSmall
	}

	switch dst.Format() {
	case SampleFormatU8:
		convertable, ok := src.(interface{ ToU8(SamplesU8) error })
		if !ok {
			return ErrConversionNotImplemented
		}
		return convertable.ToU8(dst.(SamplesU8))
	case SampleFormatI16:
		convertable, ok := src.(interface{ ToI16(SamplesI16) error })
		if !ok {
			return ErrConversionNotImplemented
		}
		return convertable.ToI16(dst.(SamplesI16))
	case SampleFormatC64:
		convertable, ok := src.(interface{ ToC64(SamplesC64) error })
		if !ok {
			return ErrConversionNotImplemented
		}
		return convertable.ToC64(dst.(SamplesC64))
	default:
		// Someone added a new type on us
		return ErrSampleFormatUnknown
	}
}

// vim: foldmethod=marker

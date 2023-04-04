// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2023
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

// LookupTable or "iq table" is a micro-optimization for extremely hotpath code
// where you make a memory / one-time CPU tradeoff for many expensive
// operations on an int8 or uint8. Since both are [2]int8 or [2]uint8,
// it's very possible to pre-compute all possible input IQ samples, since
// both could "just" be thought of as an int16, which is less than a fraction
// of a second to precompute (in IQ terms - 65535 (int16 max) is only
// 0.03 of a second at 2Msps. That being said -- this isn't free and shouldn't
// be overused.
type LookupTable interface {
	// Lookup - uncreatively - looks up the values from the source ('src')
	// IQ buffer, and writes the precomputed value to the destination ('dst')
	// buffer.
	//
	// 'dst' and 'src' MUST match the configured sample format(s).
	Lookup(dst, src Samples) (int, error)

	// SourceSampleFormat is the sample format of the precomputed table keys.
	// this must be one of SampleFormatI8 or SampleFormatU8, depending on the
	// configuration of the LookupTable.
	SourceSampleFormat() SampleFormat

	// DestinationSampleFormat is the sample format of the precomputed table
	// values. This can be any IQ type.
	DestinationSampleFormat() SampleFormat
}

// LookupTableIndexU8 will return the index into the LookupTable for an uint8
// iq sample.
func LookupTableIndexU8(v [2]uint8) uint16 {
	return *(*uint16)(unsafe.Pointer(&v[0]))
}

// LookupTableIndexI8 will return the index into the LookupTable for an int8
// iq sample.
func LookupTableIndexI8(v [2]int8) uint16 {
	return *(*uint16)(unsafe.Pointer(&v[0]))
}

// LookupTableIdentityU8 will return an identity SamplesU8 table, from 0 to max in
// index order. This can be used to apply a transform to (like Add, or
// Multiply).
func LookupTableIdentityU8() SamplesU8 {
	ret := make(SamplesU8, 65536)
	for i := range ret {
		i16 := uint16(i)
		v := *(*[2]uint8)(unsafe.Pointer(&i16))
		ret[LookupTableIndexU8(v)] = v
	}
	return ret
}

// LookupTableIdentityI8 will return an identity SamplesI8 table, from 0 to max in
// index order. This can be used to apply a transform to (like Add, or
// Multiply).
func LookupTableIdentityI8() SamplesI8 {
	ret := make(SamplesI8, 65536)
	for i := range ret {
		i16 := uint16(i)
		v := *(*[2]int8)(unsafe.Pointer(&i16))
		ret[LookupTableIndexI8(v)] = v
	}
	return ret
}

// NewLookupTable will create a new LookupTable. The 'inputFormat' is the format
// of input IQ samples. This must be either SampleFormatI8 or SampleFormatU8.
//
// On the other end, the 'lookup' buffer is the data to place into the output
// buffer depending on the input samples. The 'lookup' buffer must be exactly
// 65536 samples long.
func NewLookupTable(inputFormat SampleFormat, lookup Samples) (LookupTable, error) {
	tab, err := MakeSamples(lookup.Format(), lookup.Length())
	if err != nil {
		return nil, err
	}
	n, err := CopySamples(tab, lookup)
	if err != nil {
		return nil, err
	}
	if n != 65536 {
		return nil, fmt.Errorf("sdr: NewLookupTable requires 'lookup' be exactly 65536 samples long")
	}

	switch inputFormat {
	case SampleFormatI8, SampleFormatU8:
		break
	default:
		return nil, ErrSampleFormatUnknown
	}

	return &lookupTable{
		tab: tab,
		sf:  inputFormat,
	}, nil
}

type lookupTable struct {
	sf  SampleFormat
	tab Samples
}

func (lt *lookupTable) Lookup(dst, src Samples) (int, error) {
	if dst.Format() != lt.tab.Format() {
		return 0, ErrSampleFormatMismatch
	}
	if dst.Length() < src.Length() {
		return 0, ErrDstTooSmall
	}

	switch lt.sf {
	case SampleFormatU8:
		return lookupTableU8(dst, lt.tab, src.(SamplesU8))
	case SampleFormatI8:
		return lookupTableI8(dst, lt.tab, src.(SamplesI8))
	default:
		return 0, ErrSampleFormatMismatch
	}

	return 0, nil
}

func (lt *lookupTable) SourceSampleFormat() SampleFormat {
	return lt.sf
}

func (lt *lookupTable) DestinationSampleFormat() SampleFormat {
	return lt.tab.Format()
}

// uint8 lookup routines

func lookupTableU8(dst, tab Samples, src SamplesU8) (int, error) {
	if tab.Format() != dst.Format() {
		return 0, ErrSampleFormatMismatch
	}
	switch dst.Format() {
	case SampleFormatU8:
		return lookupTableU8ToU8(dst.(SamplesU8), tab.(SamplesU8), src)
	case SampleFormatI8:
		return lookupTableU8ToI8(dst.(SamplesI8), tab.(SamplesI8), src)
	case SampleFormatI16:
		return lookupTableU8ToI16(dst.(SamplesI16), tab.(SamplesI16), src)
	case SampleFormatC64:
		return lookupTableU8ToC64(dst.(SamplesC64), tab.(SamplesC64), src)
	default:
		return 0, ErrSampleFormatUnknown
	}
}

func lookupTableU8ToU8(dst, tab, src SamplesU8) (int, error) {
	for i, iq := range src {
		dst[i] = tab[LookupTableIndexU8(iq)]
	}
	return src.Length(), nil
}

func lookupTableU8ToI8(dst, tab SamplesI8, src SamplesU8) (int, error) {
	for i, iq := range src {
		dst[i] = tab[LookupTableIndexU8(iq)]
	}
	return src.Length(), nil
}

func lookupTableU8ToI16(dst, tab SamplesI16, src SamplesU8) (int, error) {
	for i, iq := range src {
		dst[i] = tab[LookupTableIndexU8(iq)]
	}
	return src.Length(), nil
}

func lookupTableU8ToC64(dst, tab SamplesC64, src SamplesU8) (int, error) {
	for i, iq := range src {
		dst[i] = tab[LookupTableIndexU8(iq)]
	}
	return src.Length(), nil
}

// int8 lookup routines

func lookupTableI8(dst, tab Samples, src SamplesI8) (int, error) {
	if tab.Format() != dst.Format() {
		return 0, ErrSampleFormatMismatch
	}
	switch dst.Format() {
	case SampleFormatU8:
		return lookupTableI8ToU8(dst.(SamplesU8), tab.(SamplesU8), src)
	case SampleFormatI8:
		return lookupTableI8ToI8(dst.(SamplesI8), tab.(SamplesI8), src)
	case SampleFormatI16:
		return lookupTableI8ToI16(dst.(SamplesI16), tab.(SamplesI16), src)
	case SampleFormatC64:
		return lookupTableI8ToC64(dst.(SamplesC64), tab.(SamplesC64), src)
	default:
		return 0, ErrSampleFormatUnknown
	}
}

func lookupTableI8ToU8(dst, tab SamplesU8, src SamplesI8) (int, error) {
	for i, iq := range src {
		dst[i] = tab[LookupTableIndexI8(iq)]
	}
	return src.Length(), nil
}

func lookupTableI8ToI8(dst, tab, src SamplesI8) (int, error) {
	for i, iq := range src {
		dst[i] = tab[LookupTableIndexI8(iq)]
	}
	return src.Length(), nil
}

func lookupTableI8ToI16(dst, tab SamplesI16, src SamplesI8) (int, error) {
	for i, iq := range src {
		dst[i] = tab[LookupTableIndexI8(iq)]
	}
	return src.Length(), nil
}

func lookupTableI8ToC64(dst, tab SamplesC64, src SamplesI8) (int, error) {
	for i, iq := range src {
		dst[i] = tab[LookupTableIndexI8(iq)]
	}
	return src.Length(), nil
}

// vim: foldmethod=marker

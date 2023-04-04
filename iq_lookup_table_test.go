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

package sdr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/sdr"
)

func TestLookupTableU8(t *testing.T) {
	// TODO: check the output of the Lookup calls against what we expect,
	// and ensure it's right. I left that stubbed out, but it really needs
	// to get done.
	var (
		utab = sdr.LookupTableIdentityU8()
		ctab = make(sdr.SamplesC64, utab.Length())

		uref = make(sdr.SamplesU8, 1024*32)
	)
	sdr.ConvertBuffer(ctab, utab)
	ctab.Multiply(0 + 1i)
	// utab: original uint8 buffer
	// ctab: rotated complex64 samples

	var counter uint16
	for i := range uref {
		uref[i] = [2]uint8{
			uint8(counter & 0xFF),
			uint8(int(counter & 0xFF00 >> 8)),
		}
		counter++
	}

	t.Run("U8", func(t *testing.T) {
		ltab := make(sdr.SamplesU8, ctab.Length())
		sdr.ConvertBuffer(ltab, ctab)
		tab, err := sdr.NewLookupTable(sdr.SampleFormatU8, ltab)
		assert.NoError(t, err)

		lref := make(sdr.SamplesU8, uref.Length())
		oref := make(sdr.SamplesU8, uref.Length())
		n, err := tab.Lookup(oref, lref)
		assert.NoError(t, err)
		assert.Equal(t, n, lref.Length())
	})

	t.Run("I8", func(t *testing.T) {
		ltab := make(sdr.SamplesI8, ctab.Length())
		sdr.ConvertBuffer(ltab, ctab)
		tab, err := sdr.NewLookupTable(sdr.SampleFormatU8, ltab)
		assert.NoError(t, err)

		lref := make(sdr.SamplesU8, uref.Length())
		oref := make(sdr.SamplesI8, uref.Length())
		n, err := tab.Lookup(oref, lref)
		assert.NoError(t, err)
		assert.Equal(t, n, lref.Length())
	})

	t.Run("I16", func(t *testing.T) {
		ltab := make(sdr.SamplesI16, ctab.Length())
		sdr.ConvertBuffer(ltab, ctab)
		tab, err := sdr.NewLookupTable(sdr.SampleFormatU8, ltab)
		assert.NoError(t, err)

		lref := make(sdr.SamplesU8, uref.Length())
		oref := make(sdr.SamplesI16, uref.Length())
		n, err := tab.Lookup(oref, lref)
		assert.NoError(t, err)
		assert.Equal(t, n, lref.Length())
	})

	t.Run("C64", func(t *testing.T) {
		ltab := make(sdr.SamplesC64, ctab.Length())
		sdr.ConvertBuffer(ltab, ctab)
		tab, err := sdr.NewLookupTable(sdr.SampleFormatU8, ltab)
		assert.NoError(t, err)

		lref := make(sdr.SamplesU8, uref.Length())
		oref := make(sdr.SamplesC64, uref.Length())
		n, err := tab.Lookup(oref, lref)
		assert.NoError(t, err)
		assert.Equal(t, n, lref.Length())
	})
}

func TestLookupTableI8(t *testing.T) {
	// TODO: check the output of the Lookup calls against what we expect,
	// and ensure it's right. I left that stubbed out, but it really needs
	// to get done.
	var (
		utab = sdr.LookupTableIdentityI8()
		ctab = make(sdr.SamplesC64, utab.Length())

		uref = make(sdr.SamplesI8, 1024*32)
	)
	sdr.ConvertBuffer(ctab, utab)
	ctab.Multiply(0 + 1i)
	// utab: original uint8 buffer
	// ctab: rotated complex64 samples

	var counter uint16
	for i := range uref {
		uref[i] = [2]int8{
			int8(counter & 0xFF),
			int8(int(counter&0xFF00>>8) - 127),
		}
		counter++
	}

	t.Run("U8", func(t *testing.T) {
		ltab := make(sdr.SamplesU8, ctab.Length())
		sdr.ConvertBuffer(ltab, ctab)
		tab, err := sdr.NewLookupTable(sdr.SampleFormatI8, ltab)
		assert.NoError(t, err)

		lref := make(sdr.SamplesI8, uref.Length())
		oref := make(sdr.SamplesU8, uref.Length())
		n, err := tab.Lookup(oref, lref)
		assert.NoError(t, err)
		assert.Equal(t, n, lref.Length())
	})

	t.Run("I8", func(t *testing.T) {
		ltab := make(sdr.SamplesI8, ctab.Length())
		sdr.ConvertBuffer(ltab, ctab)
		tab, err := sdr.NewLookupTable(sdr.SampleFormatI8, ltab)
		assert.NoError(t, err)

		lref := make(sdr.SamplesI8, uref.Length())
		oref := make(sdr.SamplesI8, uref.Length())
		n, err := tab.Lookup(oref, lref)
		assert.NoError(t, err)
		assert.Equal(t, n, lref.Length())
	})

	t.Run("I16", func(t *testing.T) {
		ltab := make(sdr.SamplesI16, ctab.Length())
		sdr.ConvertBuffer(ltab, ctab)
		tab, err := sdr.NewLookupTable(sdr.SampleFormatI8, ltab)
		assert.NoError(t, err)

		lref := make(sdr.SamplesI8, uref.Length())
		oref := make(sdr.SamplesI16, uref.Length())
		n, err := tab.Lookup(oref, lref)
		assert.NoError(t, err)
		assert.Equal(t, n, lref.Length())
	})

	t.Run("C64", func(t *testing.T) {
		ltab := make(sdr.SamplesC64, ctab.Length())
		sdr.ConvertBuffer(ltab, ctab)
		tab, err := sdr.NewLookupTable(sdr.SampleFormatI8, ltab)
		assert.NoError(t, err)

		lref := make(sdr.SamplesI8, uref.Length())
		oref := make(sdr.SamplesC64, uref.Length())
		n, err := tab.Lookup(oref, lref)
		assert.NoError(t, err)
		assert.Equal(t, n, lref.Length())
	})
}

// vim: foldmethod=marker

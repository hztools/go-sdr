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

package kerberos

import (
	"fmt"

	"hz.tools/sdr"
	"hz.tools/sdr/fft"
	"hz.tools/sdr/rtl/kerberos/internal"
	"hz.tools/sdr/stream"
)

// CoherentReadCloser is a slice of ReadClosers, which are in sample lock.
type CoherentReadCloser []sdr.ReadCloser

// Readers will return the ReadClosers as a Reader slice.
func (cr CoherentReadCloser) Readers() []sdr.Reader {
	ret := make([]sdr.Reader, len(cr))
	for i := range cr {
		ret[i] = cr[i]
	}
	return ret
}

// ReadersC64 will return a Reader slice, but converting to C64 while
// writing them out.
func (cr CoherentReadCloser) ReadersC64() ([]sdr.Reader, error) {
	var (
		err error
		ret = make([]sdr.Reader, len(cr))
	)
	for i := range cr {
		ret[i], err = stream.ConvertReader(cr[i], sdr.SampleFormatC64)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil

}

// Sync will check the algnment of the buffers. For best results the RNG needs
// to be on.
func (cr CoherentReadCloser) Sync(planner fft.Planner) error {
	readers, err := cr.ReadersC64()
	if err != nil {
		return err
	}
	return internal.AlignReaders(planner, readers)
}

// Close will close all the ReadClosers.
func (cr CoherentReadCloser) Close() error {
	for _, rc := range cr {
		if err := rc.Close(); err != nil {
			return err
		}
	}
	return nil
}

// StartCoherentRx will start all the RTL dongles, align the Readers, and
// return a slice of CoherentReadCloser objects.
//
// This will toggle the BiasT feature (RNG), and also flip the AGC on, and
// then set it to the passed `agc` value.
func (k Sdr) StartCoherentRx(planner fft.Planner, agc bool) (CoherentReadCloser, error) {
	var (
		err error
		sps uint
		ret = make(CoherentReadCloser, len(k))
	)

	sps, err = k[0].GetSampleRate()
	if err != nil {
		return nil, err
	}
	for i := 1; i < len(k); i++ {
		nSps, err := k[i].GetSampleRate()
		if err != nil {
			return nil, err
		}
		if nSps != sps {
			return nil, fmt.Errorf("kerberos: samples per second aren't the same")
		}
	}

	for i := range k {
		ret[i], err = k[i].StartRx()
		if err != nil {
			return nil, err
		}
	}

	if err := k.SetAutomaticGain(true); err != nil {
		ret.Close()
		return nil, err
	}

	if err := k.SetBiasT(true); err != nil {
		ret.Close()
		return nil, err
	}

	if err := ret.Sync(planner); err != nil {
		ret.Close()
		return nil, err
	}

	if err := k.SetBiasT(false); err != nil {
		ret.Close()
		return nil, err
	}

	if err := k.SetAutomaticGain(agc); err != nil {
		ret.Close()
		return nil, err
	}

	return ret, nil
}

// vim: foldmethod=marker

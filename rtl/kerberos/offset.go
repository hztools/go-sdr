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
	"log"

	"hz.tools/rf"
	"hz.tools/sdr"
	"hz.tools/sdr/fft"
	"hz.tools/sdr/rtl/kerberos/internal"
	"hz.tools/sdr/stream"
)

// OffsetSdr will return a "meta-sdr" that tunes each of the 4 Kerberos SDR
// dongles to adjacent frequency bands, and combine those SDRs into a single
// SDR at 4 times the sample rate.
type OffsetSdr struct {
	*CoherentSdr

	planner    fft.Planner
	centerFreq rf.Hz
}

// NewOffset will create a new OffsetSdr.
func NewOffset(planner fft.Planner, i1, i2, i3, i4 uint, windowSize uint) (*OffsetSdr, error) {
	sdr, err := NewCoherent(planner, i1, i2, i3, i4, windowSize)
	if err != nil {
		return nil, err
	}
	return &OffsetSdr{
		CoherentSdr: sdr,
		planner:     planner,
		centerFreq:  rf.Hz(0),
	}, nil
}

// SetSampleRate implements the sdr.Sdr interface.
func (k *OffsetSdr) SetSampleRate(sps uint) error {
	if int(sps)%len(k.Sdr) != 0 {
		return fmt.Errorf("rtl/kerberos: SampleRate is not divisible by 4")
	}
	indSps := sps / 4

	log.Printf("SampleRate set to %d - (%d per SDR)\n", sps, indSps)

	for _, s := range k.Sdr {
		if err := s.SetSampleRate(indSps); err != nil {
			return err
		}
	}

	if k.centerFreq != rf.Hz(0) {
		// the windows are determined by the sample rate, so we need
		// to re-center.
		log.Printf("Re-centering all 4 to %s\n", k.centerFreq)
		return k.SetCenterFrequency(k.centerFreq)
	}
	return nil
}

// GetSampleRate implements the sdr.Sdr interface.
func (k *OffsetSdr) GetSampleRate() (uint, error) {
	sps, err := k.Sdr[0].GetSampleRate()
	if err != nil {
		return 0, err
	}
	for i := 1; i < len(k.Sdr); i++ {
		cmpSps, err := k.Sdr[i].GetSampleRate()
		if err != nil {
			return 0, err
		}
		if cmpSps != sps {
			return 0, fmt.Errorf("rtl/kerberos: SampleRate is misaligned")
		}
	}

	totalSps := sps * uint(len(k.Sdr))
	log.Printf("Sample rate of %d per tuner, total of %d\n", sps, totalSps)
	return totalSps, nil
}

// SetCenterFrequency will set the 4 RTL tuners to adjacent frequency bands
// to allow for capture at a higher sample rate.
func (k *OffsetSdr) SetCenterFrequency(freq rf.Hz) error {
	sps, err := k.Sdr[0].GetSampleRate()
	if err != nil {
		return err
	}
	bandwidth := rf.Hz(sps)
	halfBandwidth := bandwidth / 2

	/*
	 *      cntr frq
	 *          v
	 *  +---+---+---+---+
	 *  | 2 | 3 | 0 | 1 |
	 *  +---+---+---+---+
	 *        ^ - 0.5 * sps
	 *    ^ - 1.5 * sps
	 */

	freqs := [4]rf.Hz{
		freq + (halfBandwidth),
		freq + (bandwidth + halfBandwidth),
		freq - (bandwidth + halfBandwidth),
		freq - (halfBandwidth),
	}

	log.Printf("Freq table: %s\n", freqs)

	for i := range k.Sdr {
		if err := k.Sdr[i].SetCenterFrequency(freqs[i]); err != nil {
			return err
		}
	}
	k.centerFreq = freq
	return nil
}

type readCloserCloser struct {
	sdr.ReadCloser
	closer func() error
}

func (rcc readCloserCloser) Close() error {
	if err := rcc.closer(); err != nil {
		return err
	}
	return rcc.ReadCloser.Close()
}

// StartRx will start a coherent receive, align the buffers, and return a
// reader which will stream at 4 times the sample rate of the underlying
// SDR objects.
//
// Mismatched sample rates, changing frequencies under the hood or changing
// things manually during the RX may result in some seriously weird shit.
func (k *OffsetSdr) StartRx() (sdr.ReadCloser, error) {
	readClosers, err := k.CoherentSdr.StartCoherentRx()
	if err != nil {
		return nil, err
	}

	readers := make([]sdr.Reader, len(readClosers))
	for i := range readers {
		readers[i], err = stream.ConvertReader(readClosers[i], sdr.SampleFormatC64)
		if err != nil {
			return nil, err
		}
	}

	readCloser, err := internal.GraftReaders(k.planner, readers)
	if err != nil {
		return nil, err
	}

	return &readCloserCloser{
		ReadCloser: readCloser,
		closer: func() error {
			return readClosers.Close()
		},
	}, nil
}

// vim: foldmethod=marker

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

package uhd

// #cgo pkg-config: uhd
//
// #include <uhd.h>
import "C"

import (
	"hz.tools/sdr"
)

type rxGainStage struct {
	name      string
	stageType sdr.GainStageType

	minGain float32
	maxGain float32
}

func (g rxGainStage) Range() [2]float32 {
	return [2]float32{g.minGain, g.maxGain}
}

func (g rxGainStage) Type() sdr.GainStageType {
	return g.stageType
}

func (g rxGainStage) String() string {
	return g.name
}

// SetAutomaticGain implements the sdr.Sdr interface.
func (s *Sdr) SetAutomaticGain(bool) error {
	return sdr.ErrNotSupported
}

func getRxGainStageNames(handle *C.uhd_usrp_handle, channel C.size_t) ([]string, error) {
	var (
		ret   []string
		names C.uhd_string_vector_handle
		buf   [256]C.char
		blen  = 256
	)

	if err := rvToError(C.uhd_string_vector_make(&names)); err != nil {
		return nil, err
	}
	defer C.uhd_string_vector_free(&names)

	if err := rvToError(C.uhd_usrp_get_rx_gain_names(*handle, channel, &names)); err != nil {
		return nil, err
	}

	var vlen C.size_t
	if err := rvToError(C.uhd_string_vector_size(names, &vlen)); err != nil {
		return nil, err
	}

	for i := 0; i < int(vlen); i++ {
		if err := rvToError(C.uhd_string_vector_at(names, C.size_t(i), &buf[0], C.size_t(blen))); err != nil {
			return nil, err
		}
		name := C.GoString(&buf[0])
		ret = append(ret, name)
	}
	return ret, nil
}

// GetGainStages implements the sdr.Sdr interface.
func (s *Sdr) GetGainStages() (sdr.GainStages, error) {
	var (
		ret       sdr.GainStages
		gainRange C.uhd_meta_range_handle
		start     C.double
		end       C.double
	)

	if err := rvToError(C.uhd_meta_range_make(&gainRange)); err != nil {
		return nil, err
	}
	defer C.uhd_meta_range_free(&gainRange)

	rxGainStageNames, err := getRxGainStageNames(s.handle, C.size_t(s.rxChannel))
	if err != nil {
		return nil, err
	}

	for _, gainStageName := range rxGainStageNames {
		if err := rvToError(C.uhd_usrp_get_rx_gain_range(
			*s.handle,
			C.CString(gainStageName),
			C.size_t(s.rxChannel),
			gainRange,
		)); err != nil {
			return nil, err
		}

		if err := rvToError(C.uhd_meta_range_start(gainRange, &start)); err != nil {
			return nil, err
		}

		if err := rvToError(C.uhd_meta_range_start(gainRange, &end)); err != nil {
			return nil, err
		}

		ret = append(ret, rxGainStage{
			stageType: sdr.GainStageTypeRecieve,
			name:      gainStageName,
			minGain:   float32(start),
			maxGain:   float32(end),
		})
	}

	return ret, nil
}

// GetGain implements the sdr.Sdr interface.
func (s *Sdr) GetGain(sdr.GainStage) (float32, error) {
	return 0, sdr.ErrNotSupported
}

// SetGain implements the sdr.Sdr interface.
func (s *Sdr) SetGain(sdr.GainStage, float32) error {
	return sdr.ErrNotSupported
}

// vim: foldmethod=marker

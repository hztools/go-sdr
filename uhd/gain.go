// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2021
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
	"fmt"
	"unsafe"

	"hz.tools/sdr"
)

type gainStage struct {
	prefix    string
	name      string
	stageType sdr.GainStageType

	minGain float32
	maxGain float32
	step    float32
}

func (g gainStage) Range() [2]float32 {
	return [2]float32{g.minGain, g.maxGain}
}

func (g gainStage) Type() sdr.GainStageType {
	return g.stageType
}

func (g gainStage) String() string {
	return fmt.Sprintf("%s%s", g.prefix, g.name)
}

type rxGainStage struct {
	gainStage
}

type txGainStage struct {
	gainStage
}

// SetAutomaticGain implements the sdr.Sdr interface.
func (s *Sdr) SetAutomaticGain(b bool) error {
	return rvToError(C.uhd_usrp_set_rx_agc(
		*s.handle,
		C.bool(b),
		C.size_t(s.rxChannel),
	))
}

func getStringVector(handle *C.uhd_usrp_handle, fn func(*C.uhd_string_vector_handle) error) ([]string, error) {
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

	if err := fn(&names); err != nil {
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

func getTxGainStageNames(handle *C.uhd_usrp_handle, channel C.size_t) ([]string, error) {
	return getStringVector(handle, func(names *C.uhd_string_vector_handle) error {
		return rvToError(C.uhd_usrp_get_tx_gain_names(*handle, channel, names))
	})
}

func getRxGainStageNames(handle *C.uhd_usrp_handle, channel C.size_t) ([]string, error) {
	return getStringVector(handle, func(names *C.uhd_string_vector_handle) error {
		return rvToError(C.uhd_usrp_get_rx_gain_names(*handle, channel, names))
	})
}

// GetGainStages implements the sdr.Sdr interface.
func (s *Sdr) GetGainStages() (sdr.GainStages, error) {
	var (
		ret       sdr.GainStages
		gainRange C.uhd_meta_range_handle
		start     C.double
		end       C.double
		step      C.double
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
		gsn := C.CString(gainStageName)
		err := rvToError(C.uhd_usrp_get_rx_gain_range(
			*s.handle,
			gsn,
			C.size_t(s.rxChannel),
			gainRange,
		))
		C.free(unsafe.Pointer(gsn))
		if err != nil {
			return nil, err
		}

		if err := rvToError(C.uhd_meta_range_start(gainRange, &start)); err != nil {
			return nil, err
		}

		if err := rvToError(C.uhd_meta_range_stop(gainRange, &end)); err != nil {
			return nil, err
		}

		if err := rvToError(C.uhd_meta_range_step(gainRange, &step)); err != nil {
			return nil, err
		}

		ret = append(ret, rxGainStage{gainStage{
			stageType: sdr.GainStageTypeRecieve,
			prefix:    "RX",
			name:      gainStageName,
			minGain:   float32(start),
			maxGain:   float32(end),
			step:      float32(step),
		}})
	}

	txGainStageNames, err := getTxGainStageNames(s.handle, C.size_t(s.txChannel))
	if err != nil {
		return nil, err
	}

	for _, gainStageName := range txGainStageNames {
		gsn := C.CString(gainStageName)
		err := rvToError(C.uhd_usrp_get_tx_gain_range(
			*s.handle,
			gsn,
			C.size_t(s.txChannel),
			gainRange,
		))
		C.free(unsafe.Pointer(gsn))
		if err != nil {
			return nil, err
		}

		if err := rvToError(C.uhd_meta_range_start(gainRange, &start)); err != nil {
			return nil, err
		}

		if err := rvToError(C.uhd_meta_range_stop(gainRange, &end)); err != nil {
			return nil, err
		}

		if err := rvToError(C.uhd_meta_range_step(gainRange, &step)); err != nil {
			return nil, err
		}

		ret = append(ret, txGainStage{gainStage{
			stageType: sdr.GainStageTypeTransmit,
			prefix:    "TX",
			name:      gainStageName,
			minGain:   float32(start),
			maxGain:   float32(end),
			step:      float32(step),
		}})
	}

	return ret, nil
}

// GetGain implements the sdr.Sdr interface.
func (s *Sdr) GetGain(gs sdr.GainStage) (float32, error) {
	switch gs := gs.(type) {
	case txGainStage:
		return gs.getGain(s)
	case rxGainStage:
		return gs.getGain(s)
	default:
		return 0, fmt.Errorf("uhd: unknown gain stage: %s", gs.String())
	}
}

func (g rxGainStage) getGain(s *Sdr) (float32, error) {
	var (
		gsn  = C.CString(g.name)
		gain C.double
	)
	defer C.free(unsafe.Pointer(gsn))

	if err := rvToError(C.uhd_usrp_get_rx_gain(
		*s.handle,
		C.size_t(s.rxChannel),
		gsn,
		&gain,
	)); err != nil {
		return 0, err
	}
	return float32(gain), nil
}

func (g txGainStage) getGain(s *Sdr) (float32, error) {
	var (
		gsn  = C.CString(g.name)
		gain C.double
	)
	defer C.free(unsafe.Pointer(gsn))

	if err := rvToError(C.uhd_usrp_get_tx_gain(
		*s.handle,
		C.size_t(s.txChannel),
		gsn,
		&gain,
	)); err != nil {
		return 0, err
	}
	return float32(gain), nil
}

// SetGain implements the sdr.Sdr interface.
func (s *Sdr) SetGain(gs sdr.GainStage, gain float32) error {
	switch gs := gs.(type) {
	case txGainStage:
		return gs.setGain(s, gain)
	case rxGainStage:
		return gs.setGain(s, gain)
	default:
		return fmt.Errorf("uhd: unknown gain stage: %s", gs.String())
	}
}

func (g txGainStage) setGain(s *Sdr, gain float32) error {
	gsn := C.CString(g.name)
	defer C.free(unsafe.Pointer(gsn))
	return rvToError(C.uhd_usrp_set_tx_gain(
		*s.handle,
		C.double(gain),
		C.size_t(s.txChannel),
		gsn,
	))
}

func (g rxGainStage) setGain(s *Sdr, gain float32) error {
	gsn := C.CString(g.name)
	defer C.free(unsafe.Pointer(gsn))
	return rvToError(C.uhd_usrp_set_rx_gain(
		*s.handle,
		C.double(gain),
		C.size_t(s.rxChannel),
		gsn,
	))
}

// vim: foldmethod=marker

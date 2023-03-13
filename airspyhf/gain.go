// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2022
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

package airspyhf

// #cgo pkg-config: libairspyhf
//
// #include <airspyhf.h>
import "C"

import (
	"fmt"

	"hz.tools/sdr"
)

// SetAutomaticGain implements the sdr.Sdr interface.
func (s *Sdr) SetAutomaticGain(state bool) error {
	var v C.uint8_t
	if state {
		v = 1
	}
	if C.airspyhf_set_hf_agc(s.handle, v) != C.AIRSPYHF_SUCCESS {
		return fmt.Errorf("airspyhf.Sdr.SetAutomaticGain: failed to set automatic gain")
	}

	// if state {
	// 	C.airspyhf_set_hf_agc_threshold(s.handle, 0)
	// 	C.airspyhf_set_hf_agc_threshold(s.handle, 1)
	// }

	return nil
}

type gainStage interface {
	// SetGain will set the gain on the specified Stage.
	SetGain(*Sdr, float32) error

	// SetGain will get the gain on the specified Stage.
	GetGain(*Sdr) (float32, error)
}

// GetGain implements the sdr.Sdr interface.
func (s *Sdr) GetGain(gs sdr.GainStage) (float32, error) {
	stage, ok := gs.(gainStage)
	if !ok {
		return 0, fmt.Errorf("airspyhf.Sdr.GetGain: unknown GainStage")
	}
	return stage.GetGain(s)
}

// SetGain implements the sdr.Sdr interface.
func (s *Sdr) SetGain(gs sdr.GainStage, gain float32) error {
	stage, ok := gs.(gainStage)
	if !ok {
		return fmt.Errorf("airspyhf.Sdr.SetGain: unknown GainStage")
	}
	return stage.SetGain(s, gain)
}

// GetGainStages implements the sdr.Sdr interface.
func (s *Sdr) GetGainStages() (sdr.GainStages, error) {
	return sdr.GainStages{
		attGain(newSteppedGain("Att", []int{-48, -42, -36, -30, -24, -18, -12, -6, 0})),
		ampGain(newSteppedGain("Amp", []int{0, 6})),
	}, nil
}

// The actual sdr.GainStage implementions are below this marker. They're
// both derived from the steppedGain object, and will do their best to
// pick sensible gain values for the underlying dongle.

// steppedGain is the internal base type that we're using to define
// Tuner and IF gain, since they're both clamped to fixed values.
//
// supportedGains is assumed to be the rtl-sdr specific 10th of a dB value
// (such that 110 is 11.0 dB)
type steppedGain struct {
	Name           string
	supportedGains []int
}

func (stg steppedGain) GetGainSteps() []float32 {
	ret := []float32{}
	for _, gain := range stg.supportedGains {
		ret = append(ret, float32(gain))
	}
	return ret
}

// nearestGain will return the nearest gain step to the requested
// gain step, all in rtl gain increments.
func (stg steppedGain) nearestGain(gain int) int {
	var (
		gainStep         int
		gainStepDistance = -1
	)

	for _, gainValue := range stg.supportedGains {
		gainDistance := gain - gainValue
		if gainDistance < 0 {
			gainDistance = -gainDistance
		}
		// We have the abs distance from our step to the target Gain, and now
		// we'll check to see if we're closer or further than the current
		// distance.

		if gainDistance < gainStepDistance || gainStepDistance < 0 {
			gainStepDistance = gainDistance
			gainStep = gainValue
		}
	}

	return gainStep
}

// String implements the sdr.GainStage interface.
func (stg steppedGain) String() string {
	return stg.Name
}

// Rage implements the sdr.GainStage interface.
func (stg steppedGain) Range() [2]float32 {
	sglen := len(stg.supportedGains)
	if sglen < 2 {
		return [2]float32{0, 0}
	}
	return [2]float32{
		float32(stg.supportedGains[0]),
		float32(stg.supportedGains[sglen-1]),
	}
}

// newSteppedGain will create a new "steppedGain", which is a gain
// stage where values are clamped to the nearest gain.
func newSteppedGain(name string, supportedGains []int) steppedGain {
	return steppedGain{
		Name:           name,
		supportedGains: supportedGains,
	}
}

// attGain
type attGain steppedGain

func (ag attGain) Type() sdr.GainStageType {
	return sdr.GainStageTypeRecieve | sdr.GainStageTypeAttenuator
}

// Range implements the sdr.GainStage interface.
func (ag attGain) Range() [2]float32 {
	return steppedGain(ag).Range()
}

// Get the gain steps that we support.
func (ag attGain) GetGainSteps() []float32 {
	return steppedGain(ag).GetGainSteps()
}

// String implements the sdr.GainStage interface.
func (ag attGain) String() string {
	return steppedGain(ag).String()
}

// SetGain is used as part of the rtlGainStage interface to handle
// requests to set the gain on the Sdr dongle.
func (ag attGain) SetGain(s *Sdr, gain float32) error {
	g := C.uint8_t(steppedGain(ag).nearestGain(int(gain)) / 6)
	if C.airspyhf_set_hf_att(s.handle, g) != C.AIRSPYHF_SUCCESS {
		return fmt.Errorf("airspyhf.Sdr.SetGain: set_hf_att returned an error")
	}
	return nil
}

// GetGain is used as part of the rtlGainStage interface to handle
// requests to get the gain on the Sdr dongle.
func (ag attGain) GetGain(s *Sdr) (float32, error) {
	return 0, sdr.ErrNotSupported
}

// ampGain
type ampGain steppedGain

func (ag ampGain) Type() sdr.GainStageType {
	return sdr.GainStageTypeRecieve | sdr.GainStageTypeAmp
}

// Range implements the sdr.GainStage interface.
func (ag ampGain) Range() [2]float32 {
	return steppedGain(ag).Range()
}

// Get the gain steps that we support.
func (ag ampGain) GetGainSteps() []float32 {
	return steppedGain(ag).GetGainSteps()
}

// String implements the sdr.GainStage interface.
func (ag ampGain) String() string {
	return steppedGain(ag).String()
}

// SetGain is used as part of the rtlGainStage interface to handle
// requests to set the gain on the Sdr dongle.
func (ag ampGain) SetGain(s *Sdr, gain float32) error {
	var state C.uint8_t
	if steppedGain(ag).nearestGain(int(gain)) != 0 {
		state = 1
	}
	if C.airspyhf_set_hf_lna(s.handle, state) != C.AIRSPYHF_SUCCESS {
		return fmt.Errorf("airspyhf.Sdr.SetGain: set_hf_lna returned an error")
	}
	return nil
}

// GetGain is used as part of the rtlGainStage interface to handle
// requests to get the gain on the Sdr dongle.
func (ag ampGain) GetGain(s *Sdr) (float32, error) {
	return 0, sdr.ErrNotSupported
}

// vim: foldmethod=marker

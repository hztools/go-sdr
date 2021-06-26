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

package hackrf

// #cgo pkg-config: libhackrf
//
// #include <libhackrf/hackrf.h>
import "C"

import (
	"fmt"

	"hz.tools/sdr"
)

// GetGainStages implements the sdr.Sdr interface.
func (s *Sdr) GetGainStages() (sdr.GainStages, error) {
	ret := sdr.GainStages{
		// TODO(paultag): is 14 right?
		ampGain(newSteppedGain(
			"Amp",
			sdr.GainStageTypeRecieve|sdr.GainStageTypeTransmit|sdr.GainStageTypeFE|sdr.GainStageTypeAmp,
			0, 14, 14,
		)),
		ifGain(newSteppedGain("RXIF", sdr.GainStageTypeRecieve|sdr.GainStageTypeIF, 0, 40, 8)),
		vgaRxGain(newSteppedGain("RXVGA", sdr.GainStageTypeRecieve|sdr.GainStageTypeBB, 0, 62, 2)),
		vgaTxGain(newSteppedGain("TXVGA", sdr.GainStageTypeTransmit|sdr.GainStageTypeBB, 0, 47, 1)),
	}

	return ret, nil
}

// GetGain implements the sdr.Sdr interface.
func (s *Sdr) GetGain(stage sdr.GainStage) (float32, error) {
	return 0, sdr.ErrNotSupported
}

// SetGain implements the sdr.Sdr interface.
func (s *Sdr) SetGain(gainStage sdr.GainStage, gain float32) error {
	stage, ok := gainStage.(hackrfGainStage)
	if !ok {
		return fmt.Errorf("hackrf: unknown GainStage")
	}
	return stage.SetGain(s, gain)
}

// hackrfGainStage is implemented by both gain stages in order to set the
// Gain on the HackRF
type hackrfGainStage interface {
	// SetGain will set the gain on the specified Stage.
	SetGain(*Sdr, float32) error

	// SetGain will get the gain on the specified Stage.
	GetGain(*Sdr) (float32, error)
}

//
// The actual sdr.GainStage implementions are below this marker. They're
// both derived from the steppedGain object, and will do their best to
// pick sensible gain values for the underlying dongle.
//

// steppedGain is the internal base type that we're using to define
// Tuner and IF gain, since they're both clamped to fixed values.
type steppedGain struct {
	Name           string
	supportedGains []uint32
	gainType       sdr.GainStageType
}

func newSteppedGain(name string, gainType sdr.GainStageType, low, high, step uint32) steppedGain {
	gains := []uint32{}
	var i uint32 = low
	for ; i <= high; i = i + step {
		gains = append(gains, i)
	}
	return steppedGain{
		Name:           name,
		supportedGains: gains,
		gainType:       gainType,
	}
}

// GetGainSteps implements the unspoken "steppable" Gain interface.
func (stg steppedGain) GetGainSteps() []float32 {
	ret := []float32{}
	for _, gain := range stg.supportedGains {
		ret = append(ret, float32(gain))
	}
	return ret
}

// nearestGain will return the nearest gain step to the requested
// gain step, all in hackrf gain increments.
func (stg steppedGain) nearestGain(fGain float32) uint32 {
	var (
		gainStep         uint32 = 0
		gainStepDistance int32  = -1
	)

	// TODO(paultag): use math.Round
	gain := int32(fGain)

	for _, gainValue := range stg.supportedGains {
		gainDistance := gain - int32(gainValue)
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

// Type implements the sdr.GainStage interface.
func (stg steppedGain) Type() sdr.GainStageType {
	return stg.gainType
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

// vgaRxGain
type vgaRxGain steppedGain

// Type implements the sdr.GainStage interface.
func (tg vgaRxGain) Type() sdr.GainStageType {
	return steppedGain(tg).Type()
}

// Range implements the sdr.GainStage interface.
func (tg vgaRxGain) Range() [2]float32 {
	return steppedGain(tg).Range()
}

// Get the gain steps that we support.
func (tg vgaRxGain) GetGainSteps() []float32 {
	return steppedGain(tg).GetGainSteps()
}

// String implements the sdr.GainStage interface.
func (tg vgaRxGain) String() string {
	return steppedGain(tg).String()
}

// SetGain implements the internal hackrfGain interface.
func (tg vgaRxGain) SetGain(s *Sdr, gain float32) error {
	hackrfGain := C.uint32_t(steppedGain(tg).nearestGain(gain))
	return rvToErr(C.hackrf_set_vga_gain(s.dev, hackrfGain))
}

// GetGain implements the internal hackrfGain interface.
func (tg vgaRxGain) GetGain(s *Sdr) (float32, error) {
	return 0, sdr.ErrNotSupported
}

// vgaTxGain
type vgaTxGain steppedGain

// Type implements the sdr.GainStage interface.
func (tg vgaTxGain) Type() sdr.GainStageType {
	return steppedGain(tg).Type()
}

// Range implements the sdr.GainStage interface.
func (tg vgaTxGain) Range() [2]float32 {
	return steppedGain(tg).Range()
}

// Get the gain steps that we support.
func (tg vgaTxGain) GetGainSteps() []float32 {
	return steppedGain(tg).GetGainSteps()
}

// String implements the sdr.GainStage interface.
func (tg vgaTxGain) String() string {
	return steppedGain(tg).String()
}

// SetGain implements the internal hackrfGain interface.
func (tg vgaTxGain) SetGain(s *Sdr, gain float32) error {
	hackrfGain := C.uint32_t(steppedGain(tg).nearestGain(gain))
	return rvToErr(C.hackrf_set_txvga_gain(s.dev, hackrfGain))
}

// GetGain implements the internal hackrfGain interface.
func (tg vgaTxGain) GetGain(s *Sdr) (float32, error) {
	return 0, sdr.ErrNotSupported
}

// ifGain
type ifGain steppedGain

// Type implements the sdr.GainStage interface.
func (ig ifGain) Type() sdr.GainStageType {
	return steppedGain(ig).Type()
}

// Range implements the sdr.GainStage interface.
func (ig ifGain) Range() [2]float32 {
	return steppedGain(ig).Range()
}

// Get the gain steps that we support.
func (ig ifGain) GetGainSteps() []float32 {
	return steppedGain(ig).GetGainSteps()
}

// String implements the sdr.GainStage interface.
func (ig ifGain) String() string {
	return steppedGain(ig).String()
}

// SetGain implements the internal hackrfGain interface.
func (ig ifGain) SetGain(s *Sdr, gain float32) error {
	hackrfGain := C.uint32_t(steppedGain(ig).nearestGain(gain))
	return rvToErr(C.hackrf_set_lna_gain(s.dev, hackrfGain))
}

// GetGain implements the internal hackrfGain interface.
func (ig ifGain) GetGain(s *Sdr) (float32, error) {
	return 0, sdr.ErrNotSupported
}

// ampGain
type ampGain steppedGain

// Type implements the sdr.GainStage interface.
func (ag ampGain) Type() sdr.GainStageType {
	return steppedGain(ag).Type()
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

// SetGain implements the internal hackrfGain interface.
func (ag ampGain) SetGain(s *Sdr, gain float32) error {
	onOff := C.uint32_t(steppedGain(ag).nearestGain(gain))
	var enabledU8 C.uint8_t = 0
	if onOff != 0 {
		enabledU8 = 1
	}
	s.amp = enabledU8 == 1
	return rvToErr(C.hackrf_set_amp_enable(s.dev, enabledU8))
}

// GetGain implements the internal hackrfGain interface.
func (ag ampGain) GetGain(s *Sdr) (float32, error) {
	return 0, sdr.ErrNotSupported
}

// vim: foldmethod=marker

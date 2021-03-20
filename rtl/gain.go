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

package rtl

// #cgo pkg-config: librtlsdr
//
// #include <rtl-sdr.h>
import "C"

import (
	"fmt"

	"hz.tools/sdr"
	"hz.tools/sdr/rtl/e4k"
)

// rtlGainStage is implemented by both gain stages in order to set the
// Gain on the Rtl.
//
// Technically, someone could construct a new GainStage and pass it in
// unwittingly, but it'd have to be done without using the handle member
// (since it's not exported).
type rtlGainStage interface {
	// SetGain will set the gain on the specified Stage.
	SetGain(Sdr, float32) error

	// SetGain will get the gain on the specified Stage.
	GetGain(Sdr) (float32, error)
}

// GetGain implements the sdr.Sdr interface.
func (r Sdr) GetGain(gainStage sdr.GainStage) (float32, error) {
	stage, ok := gainStage.(rtlGainStage)
	if !ok {
		return 0, fmt.Errorf("rtl: unknown GainStage")
	}
	return stage.GetGain(r)
}

// SetGain implements the sdr.Sdr interface.
func (r Sdr) SetGain(gainStage sdr.GainStage, gain float32) error {
	stage, ok := gainStage.(rtlGainStage)
	if !ok {
		return fmt.Errorf("rtl: unknown GainStage")
	}
	return stage.SetGain(r, gain)
}

// GetGainStages implements the sdr.Sdr interface.
func (r Sdr) GetGainStages() (sdr.GainStages, error) {
	return r.Tuner().GetGainStages()
}

// SetAutomaticGain will set the tuner mode to use AGC.
func (r Sdr) SetAutomaticGain(automatic bool) error {
	// TODO(paultag): Add in agc gain mode too?

	var manual C.int = 1
	if automatic {
		manual = 0
	}
	return rvToErr(C.rtlsdr_set_tuner_gain_mode(r.handle, manual))
}

//
// The actual sdr.GainStage implementions are below this marker. They're
// both derived from the steppedGain object, and will do their best to
// pick sensible gain values for the underlying dongle.
//

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
		ret = append(ret, float32(gain)/10)
	}
	return ret
}

// nearestGain will return the nearest gain step to the requested
// gain step, all in rtl gain increments.
func (stg steppedGain) nearestGain(gain int) int {
	var (
		gainStep         int = 0
		gainStepDistance int = -1
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
		float32(stg.supportedGains[0]) / 10,
		float32(stg.supportedGains[sglen-1]) / 10,
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

// tunerGain
type tunerGain steppedGain

func (tg tunerGain) Type() sdr.GainStageType {
	return sdr.GainStageTypeBB | sdr.GainStageTypeRecieve
}

// Range implements the sdr.GainStage interface.
func (tg tunerGain) Range() [2]float32 {
	return steppedGain(tg).Range()
}

// Get the gain steps that we support.
func (tg tunerGain) GetGainSteps() []float32 {
	return steppedGain(tg).GetGainSteps()
}

// String implements the sdr.GainStage interface.
func (tg tunerGain) String() string {
	return steppedGain(tg).String()
}

// SetGain is used as part of the rtlGainStage interface to handle
// requests to set the gain on the Sdr dongle.
func (tg tunerGain) SetGain(rtl Sdr, gain float32) error {
	rtlGain := C.int(steppedGain(tg).nearestGain(int(gain * 10)))
	return rvToErr(C.rtlsdr_set_tuner_gain(rtl.handle, rtlGain))
}

// GetGain is used as part of the rtlGainStage interface to handle
// requests to get the gain on the Sdr dongle.
func (tg tunerGain) GetGain(rtl Sdr) (float32, error) {
	return float32(C.rtlsdr_get_tuner_gain(rtl.handle)) / 10, nil
}

// GetGainStages will return the GainStages for the specific Tuner. The
// Sdr.GetGainStages function will implement the sdr.Sdr interface, but
// will call this function to get the underlying gains.
//
// This does not make dynamic calls, since other libraries (such as rtltcp)
// may need the GainStage values for a Tuner without it being plugged in,
// preventing the use of rtlsdr_get_tuner_gains.
func (tuner Tuner) GetGainStages() (sdr.GainStages, error) {
	stages := sdr.GainStages{tuner.tunerGain()}
	switch tuner {
	case TunerE4000:
		ifStage, err := tuner.ifGain()
		if err != nil {
			return nil, err
		}
		stages = append(stages, ifStage)
	}
	return stages, nil
}

// rtlGains will return the list of gains. In practice this ough to be a call
// to C.rtlsdr_get_tuner_gains, but hardcoding the table here allows us to
// more gracefully create GainStage objects when we don't actually have the dongle
// plugged in (e.g. rtltcp)
func (tuner Tuner) rtlGains() []int {
	switch tuner {
	case TunerE4000:
		return []int{-10, 15, 40, 65, 90, 115, 140, 165, 190, 215, 240, 290,
			340, 420}
	case TunerFC0012:
		return []int{-99, -40, 71, 179, 192}
	case TunerFC0013:
		return []int{-99, -73, -65, -63, -60, -58, -54, 58, 61, 63, 65, 67, 68,
			70, 71, 179, 181, 182, 184, 186, 188, 191, 197}
	case TunerFC2580:
		return []int{0}
	case TunerR820T, TunerR828D:
		return []int{0, 9, 14, 27, 37, 77, 87, 125, 144, 157, 166, 197, 207,
			229, 254, 280, 297, 328, 338, 364, 372, 386, 402, 421, 434, 439,
			445, 480, 496}
	default:
		return []int{0}
	}
}

// tunerGain will create the Tuner's GainStage from the hardcoded
// list of values.
func (tuner Tuner) tunerGain() tunerGain {
	return tunerGain(newSteppedGain("Tuner", tuner.rtlGains()))
}

func (tuner Tuner) ifGain() (ifGain, error) {
	if tuner != TunerE4000 {
		return ifGain{}, sdr.ErrNotSupported
	}
	ifGains := []int{}
	for i := 3; i <= 56; i++ {
		// gain steps are in rtl units, so tenths of a dB.
		ifGains = append(ifGains, i*10)
	}
	return ifGain(newSteppedGain("IF", ifGains)), nil
}

// ifGain
type ifGain steppedGain

func (ig ifGain) Type() sdr.GainStageType {
	return sdr.GainStageTypeIF | sdr.GainStageTypeRecieve
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

// SetGain is used as part of the rtlGainStage interface to handle
// requests to set the gain on the Sdr dongle.
func (ig ifGain) SetGain(rtl Sdr, gain float32) error {
	if rtl.Tuner() != TunerE4000 {
		return sdr.ErrNotSupported
	}

	stages, err := e4k.IFGainStages(uint(gain))
	if err != nil {
		return err
	}

	for stage, stageValue := range stages {
		curValue := rtl.ifStages[stage]
		if stageValue == curValue {
			continue
		}
		stage = stage + 1

		rtl.ifStages.SetGain(uint(stage), stageValue)

		if err := rvToErr(C.rtlsdr_set_tuner_if_gain(
			rtl.handle,
			C.int(stage),
			C.int(stageValue),
		)); err != nil {
			return err
		}
	}
	return nil
}

// GetGain is used as part of the rtlGainStage interface to handle
// requests to get the gain on the Sdr dongle.
func (ig ifGain) GetGain(rtl Sdr) (float32, error) {
	return 0, sdr.ErrNotSupported
}

// vim: foldmethod=marker

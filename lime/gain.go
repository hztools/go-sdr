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

package lime

// #cgo pkg-config: LimeSuite
//
// #include <lime/LimeSuite.h>
import "C"

import (
	"fmt"

	"hz.tools/sdr"
)

// GetGainStages implements the sdr.Sdr interface.
func (s *Sdr) GetGainStages() (sdr.GainStages, error) {
	ret := sdr.GainStages{
		limeGainStage{
			name:          "RX0",
			channel:       0,
			direction:     rx,
			gainStageType: sdr.GainStageTypeRecieve,
		},
		limeGainStage{
			name:          "RX1",
			channel:       1,
			direction:     rx,
			gainStageType: sdr.GainStageTypeRecieve,
		},
		limeGainStage{
			name:          "TX0",
			channel:       0,
			direction:     tx,
			gainStageType: sdr.GainStageTypeTransmit,
		},
		limeGainStage{
			name:          "TX1",
			channel:       1,
			direction:     tx,
			gainStageType: sdr.GainStageTypeTransmit,
		},
	}
	return ret, nil
}

// GetGain implements the sdr.Sdr interface.
func (s *Sdr) GetGain(gainStage sdr.GainStage) (float32, error) {
	stage, ok := gainStage.(limeGainStage)
	if !ok {
		return 0, fmt.Errorf("lime: unknown GainStage")
	}
	return stage.GetGain(s)
}

// SetGain implements the sdr.Sdr interface.
func (s *Sdr) SetGain(gainStage sdr.GainStage, gain float32) error {
	stage, ok := gainStage.(limeGainStage)
	if !ok {
		return fmt.Errorf("lime: unknown GainStage")
	}
	return stage.SetGain(s, gain)
}

type limeGainStage struct {
	name          string
	channel       uint
	direction     direction
	gainStageType sdr.GainStageType
}

func (limeGainStage) Range() [2]float32 {
	return [2]float32{0, 73}
}

func (lgs limeGainStage) String() string {
	return lgs.name
}

func (lgs limeGainStage) Type() sdr.GainStageType {
	return lgs.gainStageType
}

func (lgs limeGainStage) GetGain(s *Sdr) (float32, error) {
	var gain C.uint
	if err := rvToErr(C.LMS_GetGaindB(
		s.devPtr(),
		lgs.direction.api(),
		C.ulong(lgs.channel),
		&gain,
	)); err != nil {
		return 0, err
	}
	return float32(gain), nil
}

func (lgs limeGainStage) SetGain(s *Sdr, gain float32) error {
	return rvToErr(C.LMS_SetGaindB(
		s.devPtr(),
		lgs.direction.api(),
		C.ulong(lgs.channel),
		C.uint(gain),
	))
}

// vim: foldmethod=marker

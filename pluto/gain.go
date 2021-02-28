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

package pluto

import (
	"fmt"
	"math"

	"hz.tools/sdr"
)

type gain struct {
	minGain  float32
	maxGain  float32
	gainStep float64

	name      string
	stageType sdr.GainStageType
}

func (g gain) Range() [2]float32 {
	return [2]float32{g.minGain, g.maxGain}
}

func (g gain) Type() sdr.GainStageType {
	return g.stageType
}

func (g gain) String() string {
	return g.name
}

func (g gain) Clamp(v float64) (float64, error) {
	switch {
	case float32(v) < g.minGain:
		return 0, fmt.Errorf(
			"pluto: gain %f is below minimum gain, %f", v,
			g.minGain,
		)
	case float32(v) > g.maxGain:
		return 0, fmt.Errorf(
			"pluto: gain %f is above maximum gain, %f", v,
			g.maxGain,
		)
	default:
		return math.Round(v/g.gainStep) * g.gainStep, nil
	}
}

var (
	rxHardwareGain = gain{
		minGain:   -1,
		maxGain:   73,
		gainStep:  1,
		name:      "RX",
		stageType: sdr.GainStageTypeBB | sdr.GainStageTypeTransmit,
	}

	txHardwareGain = gain{
		minGain:   -89.75,
		maxGain:   0,
		gainStep:  0.25,
		name:      "TX",
		stageType: sdr.GainStageTypeBB | sdr.GainStageTypeRecieve,
	}
)

// SetAutomaticGain implements the sdr.Sdr interface.
func (s *Sdr) SetAutomaticGain(autoGain bool) error {
	var gcm string = "manual"
	if autoGain {
		gcm = "slow_attack"
	}
	// TODO(paultag): Should this be both Rx and Tx? What does AGC on
	// Tx mean? Defaulting to just Rx for now.
	return s.voltage0Rx.WriteString("gain_control_mode", gcm)
}

// GetGainStages implements the sdr.Sdr interface.
func (s *Sdr) GetGainStages() (sdr.GainStages, error) {
	return sdr.GainStages{rxHardwareGain, txHardwareGain}, nil
}

// GetGain implements the sdr.Sdr interface.
func (s *Sdr) GetGain(sdr.GainStage) (float32, error) {
	return 0, sdr.ErrNotSupported
}

// SetGain implements the sdr.Sdr interface.
func (s *Sdr) SetGain(gainStage sdr.GainStage, gain float32) error {
	switch gainStage {
	case rxHardwareGain:
		gain, err := rxHardwareGain.Clamp(float64(gain))
		if err != nil {
			return err
		}
		return s.voltage0Rx.WriteFloat64("hardwaregain", gain)
	case txHardwareGain:
		gain, err := txHardwareGain.Clamp(float64(gain))
		if err != nil {
			return err
		}
		return s.voltage0Tx.WriteFloat64("hardwaregain", gain)
	default:
		return fmt.Errorf("pluto: unknown gain stage: %s", gainStage.String())
	}
}

// vim: foldmethod=marker

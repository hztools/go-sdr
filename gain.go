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

package sdr

import (
	"fmt"
	"strings"
)

// GainStageType describes what type the GainStage is.
//
// This is mostly centered around where in the chain from antenna to USB
// this particular GainStage is, as well as if it's on the rx or tx side.
type GainStageType uint16

// Is will check to see if the GainStageType is another GainStageType. Direct
// comparison won't work, since we bitmask the two together, so this is a
// helpful function to do an XOR and ==.
func (gst GainStageType) Is(gainStageType GainStageType) bool {
	return (gst & gainStageType) == gainStageType
}

// String will return a short human-readable list of type names.
func (gst GainStageType) String() string {
	attrs := []string{}

	switch gst & 0xFF00 {
	case GainStageTypeRecieve | GainStageTypeTransmit:
		attrs = append(attrs, "*X")
	case GainStageTypeRecieve:
		attrs = append(attrs, "RX")
	case GainStageTypeTransmit:
		attrs = append(attrs, "TX")
	}

	if gst.Is(GainStageTypeFE) {
		attrs = append(attrs, "FE")
	}

	if gst.Is(GainStageTypeIF) {
		attrs = append(attrs, "IF")
	}

	if gst.Is(GainStageTypeBB) {
		attrs = append(attrs, "BB")
	}

	if gst.Is(GainStageTypeAmp) {
		attrs = append(attrs, "AMP")
	}

	if gst.Is(GainStageTypeAttenuator) {
		attrs = append(attrs, "ATN")
	}

	return strings.Join(attrs, ",")
}

const (
	// GainStageTypeUnknown is provided when the GainStageType is not known.
	GainStageTypeUnknown GainStageType = 0x0000

	// GainStageTypeIF represents a GainStage where Gain is applied to the
	// signal in its Intermediate Frequency stage.
	GainStageTypeIF GainStageType = 0x0001

	// GainStageTypeBB represents a GainStage where the Gain is applied at
	// the Baseband.
	GainStageTypeBB GainStageType = 0x0002

	// GainStageTypeFE represents a GainStage before the baseband/if, in
	// the radio Frontend
	GainStageTypeFE GainStageType = 0x0004

	// GainStageTypeAmp represents a GainStage before the baseband/if, in
	// the radio Frontend
	GainStageTypeAmp GainStageType = 0x0008

	// GainStageTypeAttenuator represents a GainStage which reduces the power
	// passing through rather than amplifying it.
	GainStageTypeAttenuator GainStageType = 0x0010

	// GainStageTypeRecieve represents a GainStage on the receive path.
	GainStageTypeRecieve GainStageType = 0x0100

	// GainStageTypeTransmit represents a GainStage on the transmit path.
	GainStageTypeTransmit GainStageType = 0x0200

	// External? Other?
)

// GainStage is a step at which an adjustment can be made to the values that
// flow through that stage.
type GainStage interface {
	// GainRange expresses the max and minimum values that this Gain stage
	// can be set to. The value may be negative in the case of attenuation,
	// or positive in the case of amplification.
	Range() [2]float32

	// Type will return what type of GainStage this is, usually what part
	// of the chain from antenna to USB the gain stage exists within.
	Type() GainStageType

	// String will return a human readable name to be used when referencing this
	// stage to a user. This string should match the format "%s Gain Stage" in
	// a sensible way, something like "IF", "LNA" or "Baseband" are all good
	// examples.
	String() string
}

// GainStages is a list of GainStage objects.
type GainStages []GainStage

// First will return the first GainStage that matches the criteria, returning
// nil if no such gain stage exists.
func (gs GainStages) First(gainStageType GainStageType) GainStage {
	gainStages := gs.Filter(gainStageType)
	if len(gainStages) == 0 {
		return nil
	}
	return gainStages[0]
}

// Map will return the GainStages in a map, where the key is the value returned
// by .String().
func (gs GainStages) Map() map[string]GainStage {
	ret := map[string]GainStage{}
	for _, gainStage := range gs {
		ret[gainStage.String()] = gainStage
	}
	return ret
}

// Filter will return the GainStages that match a specific criteria.
func (gs GainStages) Filter(gainStageType GainStageType) GainStages {
	ret := GainStages{}
	for _, stage := range gs {
		if stage.Type().Is(gainStageType) {
			ret = append(ret, stage)
		}
	}
	return ret
}

// SetGainStages will set the GainStages on the device by their Name as returned
// by .String().
func SetGainStages(device Sdr, gainSettings map[string]float32) error {
	gainStages, err := device.GetGainStages()
	if err != nil {
		return err
	}
	gainStagesMap := gainStages.Map()

	for gainStageName, gainValue := range gainSettings {
		gainStage, ok := gainStagesMap[gainStageName]
		if !ok {
			return fmt.Errorf("sdr: no such GainStage.Name: %s", gainStageName)
		}
		if err := device.SetGain(gainStage, gainValue); err != nil {
			return err
		}
	}

	return nil
}

// vim: foldmethod=marker

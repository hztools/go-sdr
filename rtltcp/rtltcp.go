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

package rtltcp

import (
	"fmt"

	"hz.tools/sdr/rtl"
)

// DongleInfo represents the tuner on the Server side of the rtl-tcp
// connection.
type DongleInfo struct {
	Magic          [4]byte
	TunerType      uint32
	TunerGainCount uint32
}

// Tuner will return the rtl.Tuner of the remote rtl dongle.
func (di DongleInfo) Tuner() rtl.Tuner {
	return rtl.Tuner(di.TunerType)
}

// String will return the human readable version of the DongleInfo
func (di DongleInfo) String() string {
	return fmt.Sprintf("magic=%s tunerType=%d", di.Magic, di.TunerType)
}

// Command is the first byte of the 5-byte command, indicating the requested
// action to be taken.
type Command uint8

// String will print a human readable version of the command name.
func (c Command) String() string {
	switch c {
	case CommandSetFreq:
		return "CommandSetFreq"
	case CommandSetSampleRate:
		return "CommandSetSampleRate"
	case CommandSetGainMode:
		return "CommandSetGainMode"
	case CommandSetGain:
		return "CommandSetGain"
	case CommandSetFreqCorrection:
		return "CommandSetFreqCorrection"
	case CommandSetIFGain:
		return "CommandSetIFGain"
	case CommandSetTestMode:
		return "CommandSetTestMode"
	case CommandSetAGCMode:
		return "CommandSetAGCMode"
	case CommandSetDirectSampling:
		return "CommandSetDirectSampling"
	case CommandSetOffsetTuning:
		return "CommandSetOffsetTuning"
	case CommandSetRtlXtalFreq:
		return "CommandSetRtlXtalFreq"
	case CommandSetTunerXtalFreq:
		return "CommandSetTunerXtalFreq"
	case CommandSetTunerGainByIndex:
		return "CommandSetTunerGainByIndex"
	case CommandSetBiasTee:
		return "CommandSetBiasTee"
	default:
		return "<unknown>"
	}
}

// Request encapsulates the 5-byte command request.
type Request struct {
	Command  Command
	Argument uint32
}

const (
	// CommandSetFreq requests the server tune to the frequency in Hz
	CommandSetFreq Command = 0x01

	// CommandSetSampleRate requests the server configure the number of samples
	// per second.
	CommandSetSampleRate Command = 0x02

	// CommandSetGainMode ... ?
	CommandSetGainMode Command = 0x03

	// CommandSetGain requests the server set the tuner gain.
	CommandSetGain Command = 0x04

	// CommandSetFreqCorrection requests the server set a frequency correction.
	CommandSetFreqCorrection Command = 0x05

	// CommandSetIFGain requests the server set the IF gain.
	CommandSetIFGain Command = 0x06

	// CommandSetTestMode requests the server set test mode.
	CommandSetTestMode Command = 0x07

	// CommandSetAGCMode will control the Automatic Gain Correction digital
	// stage.
	//
	// NOTE: This is very different from CommandSetGainMode. Please don't
	// use this command.
	CommandSetAGCMode Command = 0x08

	// CommandSetDirectSampling will set direct sampling.
	CommandSetDirectSampling Command = 0x09

	// CommandSetOffsetTuning will set offset tuning.
	CommandSetOffsetTuning Command = 0x0a

	// CommandSetRtlXtalFreq will set the xtal frequency.
	CommandSetRtlXtalFreq Command = 0x0b

	// CommandSetTunerXtalFreq will set the tuner xtal frequency.
	CommandSetTunerXtalFreq Command = 0x0c

	// CommandSetTunerGainByIndex will set the tuner gain.
	CommandSetTunerGainByIndex Command = 0x0d

	// CommandSetBiasTee will set the bias tee state.
	CommandSetBiasTee Command = 0x0e
)

// vim: foldmethod=marker

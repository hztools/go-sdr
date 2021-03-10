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

	"hz.tools/rf"
)

// ErrNotSupported will be returned when an SDR does not support the feature
// requested.
var ErrNotSupported error = fmt.Errorf("sdr: feature not supported by this device")

// Sdr is the generic interface that all SDRs will expose. Since this covers
// an extensive amount of functionality, it's expected some devices will not
// support a given function. If that happens, the error that must be returned
// is an ErrNotSupported.
//
// A specific SDR may support additional functionality, so be sure to check
// the documentation of the underlying SDR implementation as well!
type Sdr interface {
	// Close will free any resources held by the SDR object, and disconnect
	// from the hardware, if applicable. After this call, it's assumed
	// that any further function calls become very undefined behavior.
	Close() error

	// SetCenterFrequency will set the center of the hardware frequency to a
	// specific frequency in Hz.
	SetCenterFrequency(rf.Hz) error

	// GetCenterFrequency will get the centered hardware frequency, in Hz.
	GetCenterFrequency() (rf.Hz, error)

	// SetAutomaticGain will let the SDR take care of setting the gain as
	// required.
	SetAutomaticGain(bool) error

	// GetGainStages will return all gain stages that are supported by this
	// SDR, sorted in order from the antenna backwards to the USB port.
	GetGainStages() (GainStages, error)

	// GetGain wil return the Gain set for the specific Gain stage.
	GetGain(GainStage) (float32, error)

	// SetGain will set the Gain for a specific Gain stage.
	SetGain(GainStage, float32) error

	// SetSampleRate will set the number of samples per second that this
	// device should be sending back to us. A lower number usually gives us less
	// RF bandwidth, and a higher number may result in corruption (in the case
	// of the rtl-sdr) or dropped samples (in the case of the Pluto and friends).
	SetSampleRate(uint) error

	// GetSampleRate will get the number of samples per second that this
	// device is configured to be sending back to us.
	GetSampleRate() (uint, error)

	// SampleFormat returns the type of this vector, as exported by the
	// SampleFormat enum.
	SampleFormat() SampleFormat

	// SetPPM will set the error in parts-per-million. This is used to adjust
	// for clock skew.
	SetPPM(int) error

	// HardwareInfo will return information about the connected SDR.
	HardwareInfo() HardwareInfo
}

// HardwareInfo contains information about the connected SDR.
//
// Some subset of this information may be populated, none of it is
// a hard requirement if it does not exist. Not all SDRs will have a
// Serial, for example.
type HardwareInfo struct {
	// Manufacturer is the person, company or group that created this SDR.
	Manufacturer string

	// Product is the name of the specific SDR product connected.
	Product string

	// Serial is an identifier that is unique to the connected SDR.
	Serial string
}

// Transmitter is an "extension" of the SDR Interface, it contains all the
// common control methods, plus additional bits to transmit iq data over
// the airwaves.
//
// This can either be used as part of a function signature if your code really
// only needs to transmit, or as part of a type-cast to determine if the SDR
// is capable of transmitting.
type Transmitter interface {
	Sdr

	// StartTx will begin to transmit on the configured frequency, and start to
	// stream iq samples written to the underlying hardware to be sent over
	// the air.
	//
	// It's absolutely imperitive that the producing code feed iq samples into
	// the transmitter at the specified rate, or bad things may happen and
	// cause wildly unpredictable things.
	StartTx() (WriteCloser, error)
}

// Receiver is an "extension" of the SDR Interface, it contains all the
// common control methods, plus additional bits to recieve iq data from
// the airwaves.
//
// This can either be used as part of a function signature if your code really
// only needs to recieve, or as part of a type-cast to determine if the SDR
// is capable of recieving.
type Receiver interface {
	Sdr

	// StartRx will listen on the configured frequency and start to stream iq
	/// samples to be read out of the provided Reader. It's absolutely
	// imperitive that the consuming code will actively consume from the
	// Reader, or backlogged samples can result in dropped samples or other
	// error conditions. Those error conditions are not defined at this time,
	// but may break in wildly unpredictable ways, since the time sensitive
	// SDR code may hang waiting for reads.
	StartRx() (ReadCloser, error)
}

// Transceiver is an "extension" of the SDR Interface, it contains all the
// common control methods, plus additional bits to both transmit and recieve
// iq data.
//
// This can either be used as part of a function signature if your code really
// only needs to both recieve and transmit, or as part of a type-cast to
// determine if the SDR is capable of both recieving and transmitting.
type Transceiver interface {
	Sdr
	Receiver
	Transmitter
}

// vim: foldmethod=marker

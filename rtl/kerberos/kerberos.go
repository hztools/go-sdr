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

package kerberos

import (
	"hz.tools/rf"
	"hz.tools/sdr"
	"hz.tools/sdr/debug"
	"hz.tools/sdr/rtl"
)

func init() {
	debug.RegisterRadioDriver("hz.tools/sdr/rtl/kerberos.Sdr")
}

// Sdr is a Kerberos SDR, 4 RTL-SDR dongles in one!
type Sdr [4]*rtl.Sdr

// New will create a new Kerberos SDR
func New(i1, i2, i3, i4 uint, windowSize uint) (*Sdr, error) {
	var (
		err error
		sdr = &Sdr{}
	)
	for i := range sdr {
		sdr[i], err = rtl.New(uint(i), 0)
		if err != nil {
			return nil, err
		}
	}
	return sdr, nil
}

// Close implements the sdr.Sdr interface.
func (k Sdr) Close() error {
	for _, s := range k {
		if err := s.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Tuner will return the RTL-SDR Tuner object in the Kerberos SDR provided.
func (k Sdr) Tuner() rtl.Tuner {
	return k[0].Tuner()
}

// GetCenterFrequency implements the sdr.Sdr interface.
func (k Sdr) GetCenterFrequency() (rf.Hz, error) {
	return rf.Hz(0), sdr.ErrNotSupported
}

// GetGain implements the sdr.Sdr interface.
func (k Sdr) GetGain(gainStage sdr.GainStage) (float32, error) {
	return k[0].GetGain(gainStage)
}

// GetGainStages implements the sdr.Sdr interface.
func (k Sdr) GetGainStages() (sdr.GainStages, error) {
	return k[0].GetGainStages()
}

// GetPPM gets the PPM Offset.
func (k Sdr) GetPPM() int {
	return k[0].GetPPM()
}

// HardwareInfo implements the sdr.Sdr interface.
func (k Sdr) HardwareInfo() sdr.HardwareInfo {
	// TODO(paultag): Implement me
	return sdr.HardwareInfo{}
}

// GetSampleRate implements the sdr.Sdr interface.
func (k Sdr) GetSampleRate() (uint, error) {
	return k[0].GetSampleRate()
}

// ResetBuffer will reset the rtl-sdr Buffers.
func (k Sdr) ResetBuffer() error {
	for _, s := range k {
		if err := s.ResetBuffer(); err != nil {
			return err
		}
	}
	return nil
}

// SampleFormat implements the sdr.Sdr interface.
func (k Sdr) SampleFormat() sdr.SampleFormat {
	return sdr.SampleFormatU8
}

// SetAutomaticGain implements the sdr.Sdr interface.
func (k Sdr) SetAutomaticGain(automatic bool) error {
	for _, s := range k {
		if err := s.SetAutomaticGain(automatic); err != nil {
			return err
		}
	}
	return nil
}

// SetBiasT will toggle the GPIO Pin #1, which is the RNG source on the KSDR.
func (k Sdr) SetBiasT(on bool) error {
	return k[0].SetBiasT(on)
}

// SetBiasTGPIO will toggle a GPIO Pin.
func (k Sdr) SetBiasTGPIO(pin int, on bool) error {
	return k[0].SetBiasTGPIO(pin, on)
}

// SetCenterFrequency implements the sdr.Sdr interface.
func (k Sdr) SetCenterFrequency(freq rf.Hz) error {
	for i := range k {
		if err := k[i].SetCenterFrequency(freq); err != nil {
			return err
		}
	}
	return nil
}

// SetGain implements the sdr.Sdr interface.
func (k Sdr) SetGain(gainStage sdr.GainStage, gain float32) error {
	for _, s := range k {
		if err := s.SetGain(gainStage, gain); err != nil {
			return err
		}
	}
	return nil
}

// SetPPM sets the PPM offset.
func (k Sdr) SetPPM(ppm int) error {
	for _, s := range k {
		if err := s.SetPPM(ppm); err != nil {
			return err
		}
	}
	return nil
}

// SetSampleRate implements the sdr.Sdr interface.
func (k Sdr) SetSampleRate(sps uint) error {
	for _, s := range k {
		if err := s.SetSampleRate(sps); err != nil {
			return err
		}
	}
	return nil
}

// SetTestMode will turn the RTL-SDR into a test mode.
func (k Sdr) SetTestMode(on bool) error {
	for _, s := range k {
		if err := s.SetTestMode(true); err != nil {
			return err
		}
	}
	return nil
}

// vim: foldmethod=marker

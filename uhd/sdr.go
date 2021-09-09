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

package uhd

// #cgo pkg-config: uhd
//
// #include <uhd.h>
import "C"

import (
	"hz.tools/rf"
	"hz.tools/sdr"
	"hz.tools/sdr/debug"
)

func init() {
	debug.RegisterRadioDriver("hz.tools/sdr/uhd.Sdr")
}

// Sdr is a UHD backed Software Defined Radio. This implements the sdr.Sdr
// interface.
type Sdr struct {
	handle       *C.uhd_usrp_handle
	sampleFormat sdr.SampleFormat

	rxChannel int

	sampleRate uint

	hi sdr.HardwareInfo
}

// Options contains arguments used to configure the UHD Radio.
type Options struct {
	// Args is passed to uhd_usrp_make as device arguments.
	Args string

	// TODO(paultag): Flag RX/TX caps

	// RxChannel is the channel to use for RX operations.
	RxChannel int

	// SampleFormat to be used internally.
	//
	// Currently supported types:
	//   - sdr.SampleFormatI16
	//   - sdr.SampleFormatC64
	//
	SampleFormat sdr.SampleFormat
}

// Open will connect to an USRP Radio.
func Open(opts Options) (*Sdr, error) {
	var (
		usrp C.uhd_usrp_handle

		buf  [256]C.char
		blen = 256
	)

	if err := rvToError(C.uhd_usrp_make(&usrp, C.CString(opts.Args))); err != nil {
		return nil, err
	}

	if opts.SampleFormat == 0 {
		opts.SampleFormat = sdr.SampleFormatI16
	}

	if err := rvToError(C.uhd_usrp_get_mboard_name(
		usrp,
		0,
		&buf[0],
		C.size_t(blen),
	)); err != nil {
		C.uhd_usrp_free(&usrp)
		return nil, err
	}

	mboard := C.GoString(&buf[0])

	hi := sdr.HardwareInfo{
		Manufacturer: "Ettus Research", // TODO(paultag): Fix this too
		Product:      mboard,
		Serial:       "", // TODO(paultag): Do this
	}

	return &Sdr{
		handle:       &usrp,
		sampleFormat: opts.SampleFormat,
		rxChannel:    opts.RxChannel,
		hi:           hi,
	}, nil
}

// Close will release all held handles.
func (s *Sdr) Close() error {
	return rvToError(C.uhd_usrp_free(s.handle))
}

// GetCenterFrequency implements the sdr.Sdr interface.
func (s *Sdr) GetCenterFrequency() (rf.Hz, error) {
	var freq C.double
	if err := rvToError(C.uhd_usrp_get_rx_freq(*s.handle, C.size_t(s.rxChannel), &freq)); err != nil {
		return rf.Hz(0), err
	}
	return rf.Hz(freq), nil
}

// SetCenterFrequency implements the sdr.Sdr interface.
func (s *Sdr) SetCenterFrequency(freq rf.Hz) error {
	var (
		tuneRequest C.uhd_tune_request_t
		tuneResult  C.uhd_tune_result_t
	)

	tuneRequest.target_freq = C.double(freq)
	tuneRequest.rf_freq_policy = C.UHD_TUNE_REQUEST_POLICY_AUTO
	tuneRequest.dsp_freq_policy = C.UHD_TUNE_REQUEST_POLICY_AUTO

	// TODO(paultag): set tx freq

	return rvToError(C.uhd_usrp_set_rx_freq(
		*s.handle,
		&tuneRequest,
		C.size_t(s.rxChannel),
		&tuneResult,
	))
}

// SetSampleRate implements the sdr.Sdr interface.
func (s *Sdr) SetSampleRate(rate uint) error {
	return rvToError(C.uhd_usrp_set_rx_rate(*s.handle, C.double(rate), C.size_t(s.rxChannel)))
}

// GetSampleRate implements the sdr.Sdr interface.
func (s *Sdr) GetSampleRate() (uint, error) {
	// TODO(paultag): the sample rate is returned as a float. This isn't
	// quite ideal, given that we treat it as a uint.
	var rate C.double
	if err := rvToError(C.uhd_usrp_get_rx_rate(*s.handle, C.size_t(s.rxChannel), &rate)); err != nil {
		return 0, err
	}
	return uint(rate), nil
}

// SampleFormat implements the sdr.Sdr interface.
func (s *Sdr) SampleFormat() sdr.SampleFormat {
	return s.sampleFormat
}

// SetPPM implements the sdr.Sdr interface.
func (s *Sdr) SetPPM(int) error {
	return sdr.ErrNotSupported
}

// HardwareInfo implements the sdr.Sdr interface.
func (s *Sdr) HardwareInfo() sdr.HardwareInfo {
	return s.hi
}

// vim: foldmethod=marker

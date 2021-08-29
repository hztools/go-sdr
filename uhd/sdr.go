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
)

// Sdr is a UHD backed Software Defined Radio. This implements the sdr.Sdr
// interface.
type Sdr struct {
	handle *C.uhd_usrp_handle

	rxStreamer *C.uhd_rx_streamer_handle
	rxChannel  int

	sampleRate uint
}

// Close will release all held handles.
func (s *Sdr) Close() error {
	if err := rvToError(C.uhd_rx_streamer_free(s.rxStreamer)); err != nil {
		return err
	}

	if err := rvToError(C.uhd_usrp_free(s.handle)); err != nil {
		return err
	}
	return nil
}

// GetCenterFrequency implements the sdr.Sdr interface.
func (s *Sdr) GetCenterFrequency() (rf.Hz, error) {
	var freq C.double
	if err := rvToError(C.uhd_usrp_get_rx_freq(*s.handle, C.ulong(s.rxChannel), &freq)); err != nil {
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
		C.ulong(s.rxChannel),
		&tuneResult,
	))
}

// Options contains arguments used to configure the UHD Radio.
type Options struct {
	// Args is passed to uhd_usrp_make as device arguments.
	Args string

	// TODO(paultag): Flag RX/TX caps

	// RxChannel is the channel to use for RX operations.
	RxChannel int
}

// Open will connect to an USRP Radio.
func Open(opts Options) (*Sdr, error) {
	var (
		usrp       C.uhd_usrp_handle
		rxStreamer C.uhd_rx_streamer_handle
	)

	if err := rvToError(C.uhd_usrp_make(&usrp, C.CString(opts.Args))); err != nil {
		return nil, err
	}

	if err := rvToError(C.uhd_rx_streamer_make(&rxStreamer)); err != nil {
		C.uhd_usrp_free(&usrp)
		return nil, err
	}

	return &Sdr{
		handle:     &usrp,
		rxStreamer: &rxStreamer,
		rxChannel:  opts.RxChannel,
	}, nil
}

// vim: foldmethod=marker

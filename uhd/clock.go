// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2021
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
	"time"
	"unsafe"
)

// SetTimeNextPPS will set the time at the next PPS pulse.
func (s *Sdr) SetTimeNextPPS(d time.Duration) error {
	secs, frac := splitDuration(d)

	return rvToError(C.uhd_usrp_set_time_next_pps(
		*s.handle,
		C.time_t(secs),
		C.double(frac),
		0,
	))
}

// SetTimeNow will set the offset clock to the provided Duration. This isn't
// a time.Time since it's a bit confusing at this particular i/o boundary.
func (s *Sdr) SetTimeNow(d time.Duration) error {
	secs, frac := splitDuration(d)

	// TODO(paultag): Multiple Mboards?
	return rvToError(C.uhd_usrp_set_time_now(
		*s.handle,
		C.time_t(secs),
		C.double(frac),
		0,
	))
}

// splitDuration will split a time.Duration into USRP's time format, which
// is whole seconds and fractional seconds.
func splitDuration(d time.Duration) (C.time_t, C.double) {
	secs := d.Seconds()
	frac := float64(d.Nanoseconds()) / 1e+9
	return C.time_t(secs), C.double(frac)
}

// newDuration will create a time.Duration from USRP's time format, made
// up of whole seconds and fractional seconds.
func newDuration(secs C.time_t, frac C.double) time.Duration {
	return time.Duration((float64(time.Second) * float64(frac)) +
		(float64(time.Second) * float64(secs)))
}

// GetTimeNow will get the USRP's Time information at the instant it was
// requested.
func (s *Sdr) GetTimeNow() (time.Duration, error) {
	var (
		secs C.time_t
		frac C.double
	)

	if err := rvToError(C.uhd_usrp_get_time_now(
		*s.handle,
		0,
		&secs,
		&frac,
	)); err != nil {
		return time.Duration(0), err
	}

	return newDuration(secs, frac), nil
}

// SetTimeSource will set the clock ref for the USRP.
func (s *Sdr) SetTimeSource(what string) error {
	cWhat := C.CString(what)
	defer C.free(unsafe.Pointer(cWhat))
	return rvToError(C.uhd_usrp_set_time_source(
		*s.handle,
		cWhat,
		0,
	))
}

// TODO:
//
//  - uhd_usrp_set_time_unknown_pps
//
//  - uhd_usrp_get_time_last_pps
//
//  - uhd_usrp_get_time_source
//  - uhd_usrp_get_time_synchronized
//

// vim: foldmethod=marker

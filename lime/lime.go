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

package lime

// #cgo pkg-config: LimeSuite
//
// #include <lime/LimeSuite.h>
import "C"

import (
	"fmt"
	"strings"
	"unsafe"

	"hz.tools/rf"
	"hz.tools/sdr"
	"hz.tools/sdr/internal/warning"
)

type direction bool

const (
	rx direction = false
	tx direction = true
)

func (d direction) api() C.bool {
	return C.bool(d)
}

func rvToErr(rv C.int) error {
	if rv != 0 {
		v := C.GoString(C.LMS_GetLastErrorMessage())
		return fmt.Errorf("lime: err %d: %s", rv, v)
	}
	return nil
}

// Open will open the first LimeSDR plugged into the system.
func Open() (*Sdr, error) {
	warning.Experimental("limesdr")

	var (
		device    = C.lms_device_t{}
		devicePtr = unsafe.Pointer(&device)
	)

	// TODO(paultag): Update `nil, nil` to allow specific SDR loading.
	if err := rvToErr(C.LMS_Open(&devicePtr, nil, nil)); err != nil {
		return nil, err
	}

	if err := rvToErr(C.LMS_Reset(devicePtr)); err != nil {
		return nil, err
	}

	if err := rvToErr(C.LMS_Init(devicePtr)); err != nil {
		return nil, err
	}

	// TODO(paultag): There's a lot more in here that would be nice to provide
	// to users. Maybe extend sdr.HardwareInfo ?
	devInfo := C.LMS_GetDeviceInfo(devicePtr)
	info := sdr.HardwareInfo{
		Serial:       fmt.Sprintf("%x", devInfo.boardSerialNumber),
		Product:      strings.Trim(C.GoStringN(&devInfo.deviceName[0], 32), "\x00"),
		Manufacturer: "Lime",
	}

	s := &Sdr{
		dev:  (*C.lms_device_t)(devicePtr),
		info: info,
	}

	return s, nil
}

// Sdr is a Lime SDR of some type.
type Sdr struct {
	dev  *C.lms_device_t
	info sdr.HardwareInfo

	sampleRate uint
}

func (s *Sdr) devPtr() unsafe.Pointer {
	return unsafe.Pointer(s.dev)
}

// Close implements the sdr.Sdr interface.
func (s *Sdr) Close() error {
	return rvToErr(C.LMS_Close(s.devPtr()))
}

// SetSampleRate implements the sdr.Sdr interface.
func (s *Sdr) SetSampleRate(rate uint) error {

	var (
		targetRate uint = rate
		oversample uint = 16
	)

	if err := rvToErr(C.LMS_SetSampleRate(
		s.devPtr(),
		C.double(float64(targetRate)),
		C.ulong(oversample),
	)); err != nil {
		return err
	}

	s.sampleRate = rate
	return nil
}

// GetCenterFrequency implements the sdr.Sdr interface.
func (s *Sdr) GetCenterFrequency() (rf.Hz, error) {
	return 0, sdr.ErrNotSupported
}

// GetSampleRate implements the sdr.Sdr interface.
func (s *Sdr) GetSampleRate() (uint, error) {
	return 0, sdr.ErrNotSupported
}

// HardwareInfo implements the sdr.Sdr interface.
func (s *Sdr) HardwareInfo() sdr.HardwareInfo {
	return s.info
}

// SetCenterFrequency implements the sdr.Sdr interface.
func (s *Sdr) SetCenterFrequency(r rf.Hz) error {
	// Here, we set both the RX and TX frequency at the same time, which
	// is perhaps a bug in the sdr.Sdr interface.
	//
	// This is something we do in the PlutoSDR driver too, which is, again
	// a bit awkward. It'd be nice to support mismatched RX/TX for things like
	// transmit/rx channels (or like, a repeater).
	//
	// But, for now, we can live with this.

	if err := rvToErr(C.LMS_SetLOFrequency(
		s.devPtr(),
		rx.api(),
		0,
		C.double(float64(r)),
	)); err != nil {
		return err
	}
	return rvToErr(C.LMS_SetLOFrequency(
		s.devPtr(),
		tx.api(),
		0,
		C.double(float64(r)),
	))
}

// SetAutomaticGain implements the sdr.Sdr interface.
func (s *Sdr) SetAutomaticGain(bool) error {
	return sdr.ErrNotSupported
}

// SampleFormat implements the sdr.Sdr interface.
func (s *Sdr) SampleFormat() sdr.SampleFormat {
	return sdr.SampleFormatI16
}

// SetPPM implements the sdr.Sdr interface.
func (s *Sdr) SetPPM(int) error {
	return sdr.ErrNotSupported
}

// vim: foldmethod=marker
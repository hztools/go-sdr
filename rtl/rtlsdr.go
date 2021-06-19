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

package rtl

// #cgo pkg-config: librtlsdr
//
// #include <stdint.h>
// #include <malloc.h>
//
// #include <rtl-sdr.h>
import "C"

import (
	"fmt"
	"unsafe"

	"hz.tools/rf"
	"hz.tools/sdr"
	"hz.tools/sdr/debug"
	"hz.tools/sdr/rtl/e4k"
)

func init() {
	debug.RegisterRadioDriver("hz.tools/sdr/rtl.Sdr")
}

// DeviceCount will return the number of rtlsdr devices present on the
// system.
func DeviceCount() uint {
	return uint(C.rtlsdr_get_device_count())
}

// DeviceIndexBySerial will get the device index that has the Serial provided.
func DeviceIndexBySerial(serial string) (uint, error) {
	index := C.rtlsdr_get_index_by_serial(C.CString(serial))

	// return -1 if name is NULL
	// return -2 if no devices were found at all
	// return -3 if devices were found, but none with matching name

	switch index {
	case -1:
		return 0, fmt.Errorf("rtl: name is NULL")
	case -2:
		return 0, fmt.Errorf("rtl: no devices found")
	case -3:
		return 0, fmt.Errorf("rtl: no device matching that serial found")
	}

	return uint(index), nil
}

// New will create a new Sdr struct, and initialize the internal
// handles as required.
//
// index      corresponds to the index into the number of devices (as seen by
//            DeviceCount) to open.
//
// windowSize instructs the rtlsdr library as to how many iq samples to deliver
//            per callback.
//
func New(index uint, windowSize uint) (*Sdr, error) {
	if windowSize == 0 {
		windowSize = 16 * 32 * 512
	}

	ret := Sdr{
		windowSize: windowSize,
		ifStages:   &e4k.Stages{},
	}
	if err := rvToErr(C.rtlsdr_open(&ret.handle, C.uint(index))); err != nil {
		return nil, err
	}
	inf, err := ret.info()
	if err != nil {
		ret.Close()
		return nil, err
	}
	ret.hardwareInfo = inf.HardwareInfo()
	return &ret, nil
}

// Sdr is a handle to internal rtlsdr state used by the underlying C
// library.
type Sdr struct {
	handle     *C.rtlsdr_dev_t
	windowSize uint

	ifStages     *e4k.Stages
	hardwareInfo sdr.HardwareInfo
}

// SampleFormat will return the IQ sample format. For the rtlsdr this
// is always SampleFormatU8.
func (r Sdr) SampleFormat() sdr.SampleFormat {
	return sdr.SampleFormatU8
}

// info will provide information about the Sdr that this data relates
// to.
//
// TODO(paultag): Remove this now that we have sdr.HardwareInfo
//
type info struct {
	Manufacturer string
	Product      string
	Serial       string
}

func (i info) HardwareInfo() sdr.HardwareInfo {
	return sdr.HardwareInfo{
		Manufacturer: i.Manufacturer,
		Product:      i.Product,
		Serial:       i.Serial,
	}
}

// InfoByDeviceIndex will return a HardwareInfo struct by a device index.
func InfoByDeviceIndex(index uint) (*sdr.HardwareInfo, error) {
	var cMfgr *C.char = (*C.char)(C.malloc(255))
	defer C.free(unsafe.Pointer(cMfgr))

	var cProd *C.char = (*C.char)(C.malloc(255))
	defer C.free(unsafe.Pointer(cProd))

	var cSerial *C.char = (*C.char)(C.malloc(255))
	defer C.free(unsafe.Pointer(cSerial))

	if err := rvToErr(C.rtlsdr_get_device_usb_strings(C.uint(index), cMfgr, cProd, cSerial)); err != nil {
		return nil, err
	}

	i := info{
		Manufacturer: C.GoString(cMfgr),
		Product:      C.GoString(cProd),
		Serial:       C.GoString(cSerial),
	}.HardwareInfo()
	return &i, nil
}

// HardwareInfo implements the sdr.Sdr interface
func (r Sdr) HardwareInfo() sdr.HardwareInfo {
	return r.hardwareInfo
}

// info will fetch and populate info relating to the device that has
// been opened.
func (r Sdr) info() (*info, error) {
	// alloc *C.char[255], pass to the C function, turn to Go, defer the
	// free and return the info blob

	var cMfgr *C.char = (*C.char)(C.malloc(255))
	defer C.free(unsafe.Pointer(cMfgr))

	var cProd *C.char = (*C.char)(C.malloc(255))
	defer C.free(unsafe.Pointer(cProd))

	var cSerial *C.char = (*C.char)(C.malloc(255))
	defer C.free(unsafe.Pointer(cSerial))

	if err := rvToErr(C.rtlsdr_get_usb_strings(r.handle, cMfgr, cProd, cSerial)); err != nil {
		return nil, err
	}

	return &info{
		Manufacturer: C.GoString(cMfgr),
		Product:      C.GoString(cProd),
		Serial:       C.GoString(cSerial),
	}, nil

}

// Close will call the underlying rtlsdr library to close the handle that
// we opened when creating this object. This should always be called when
// finished with the sdr.
func (r Sdr) Close() error {
	return rvToErr(C.rtlsdr_close(r.handle))
}

// SetCenterFrequency will set the center frequency that the rtlsdr
// will tune to.
func (r Sdr) SetCenterFrequency(freq rf.Hz) error {
	return rvToErr(C.rtlsdr_set_center_freq(r.handle, C.uint32_t(freq)))
}

// GetCenterFrequency will return the center frequency that the rtlsdr
// is tuned to.
func (r Sdr) GetCenterFrequency() (rf.Hz, error) {
	cFreq := C.rtlsdr_get_center_freq(r.handle)
	return rf.Hz(cFreq), nil
}

// SetSampleRate will set the number of samples per second. This will not
// change the window size.
func (r Sdr) SetSampleRate(sps uint) error {
	return rvToErr(C.rtlsdr_set_sample_rate(r.handle, C.uint32_t(sps)))
}

// GetSampleRate will get the number of samples per second.
func (r Sdr) GetSampleRate() (uint, error) {
	return uint(C.rtlsdr_get_sample_rate(r.handle)), nil
}

// GetSamplesPerWindow will return the number of samples contained in one
// windows-worth of iq data.
func (r Sdr) GetSamplesPerWindow() (uint, error) {
	return r.windowSize / 2, nil
}

// ResetBuffer will reset the internal rtlsdr buffer(s).
func (r Sdr) ResetBuffer() error {
	return rvToErr(C.rtlsdr_reset_buffer(r.handle))
}

// SetPPM will set the PPM skew.
func (r Sdr) SetPPM(ppm int) error {
	return rvToErr(C.rtlsdr_set_freq_correction(r.handle, C.int(ppm)))
}

// GetPPM will get the PPM skew.
func (r Sdr) GetPPM() int {
	return int(C.rtlsdr_get_freq_correction(r.handle))
}

// SetTestMode will turn on and off test mode.
//
// Test mode will cause every byte to be the one larger than the next byte,
// and on overflow, return to 0. This is useful to detect cases where you're
// dropping packets, or to ensure that your code can fully process the data
// end-to-end in real-time.
func (r Sdr) SetTestMode(on bool) error {
	if on {
		return rvToErr(C.rtlsdr_set_testmode(r.handle, 1))
	}
	return rvToErr(C.rtlsdr_set_testmode(r.handle, 0))
}

// SetBiasT will enable or disable the bias tee.
func (r Sdr) SetBiasT(on bool) error {
	// TODO(paultag): check if return value is -1, which is uninitialized
	if on {
		return rvToErr(C.rtlsdr_set_bias_tee(r.handle, 1))
	}
	return rvToErr(C.rtlsdr_set_bias_tee(r.handle, 0))
}

// SetBiasTGPIO will enable or disable the bias tee.
func (r Sdr) SetBiasTGPIO(pin int, on bool) error {
	// TODO(paultag): check if return value is -1, which is uninitialized

	if on {
		return rvToErr(C.rtlsdr_set_bias_tee_gpio(r.handle, C.int(pin), 1))
	}
	return rvToErr(C.rtlsdr_set_bias_tee_gpio(r.handle, C.int(pin), 0))
}

// Tuner will return the rtlsdr Tuner type. This can be used to determine
// the behavior of some of the Gain options, as well as well as performance.
func (r Sdr) Tuner() Tuner {
	return Tuner(C.rtlsdr_get_tuner_type(r.handle))
}

// Tuner is an enum that represents an rtlsdr Tuner chipset.
type Tuner uint8

// String will return the human readable name for the Tuner type.
func (t Tuner) String() string {
	switch t {
	case TunerE4000:
		return "E4000"
	case TunerFC0012:
		return "FC0012"
	case TunerFC0013:
		return "FC0013"
	case TunerFC2580:
		return "FC2580"
	case TunerR820T:
		return "R820T"
	case TunerR828D:
		return "R828D"
	case TunerUnknown:
		return "Unknown"
	default:
		return "UNKNOWN"
	}
}

var (
	// TunerUnknown is used when the underlying rtlsdr Tuner is not known
	// by the underlying rtlsdr library.
	TunerUnknown Tuner = C.RTLSDR_TUNER_UNKNOWN

	// TunerE4000 represents the rtlsdr E4000 tuner type.
	TunerE4000 Tuner = C.RTLSDR_TUNER_E4000

	// TunerFC0012 represents the rtlsdr FC0012 tuner type.
	TunerFC0012 Tuner = C.RTLSDR_TUNER_FC0012

	// TunerFC0013 represents the rtlsdr FC0013 tuner type.
	TunerFC0013 Tuner = C.RTLSDR_TUNER_FC0013

	// TunerFC2580 represents the rtlsdr FC2580 tuner type.
	TunerFC2580 Tuner = C.RTLSDR_TUNER_FC2580

	// TunerR820T represents the rtlsdr R820T tuner type.
	TunerR820T Tuner = C.RTLSDR_TUNER_R820T

	// TunerR828D represents the rtlsdr R828D tuner type.
	TunerR828D Tuner = C.RTLSDR_TUNER_R828D
)

// vim: foldmethod=marker

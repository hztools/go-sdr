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

package hackrf

// #cgo pkg-config: libhackrf
//
// #include <libhackrf/hackrf.h>
import "C"

import (
	"fmt"
	"unsafe"

	"hz.tools/rf"
	"hz.tools/sdr"
	"hz.tools/sdr/debug"
)

func init() {
	debug.RegisterRadioDriver("hz.tools/sdr/hackrf.Sdr")
}

var (
	hasInit         bool = false
	usbBoardMapping      = map[uint32]Board{
		0x604B: BoardJawbreaker,
		0x6089: BoardHackRfOne,
		0xFFFF: BoardInvalid,
	}
)

func checkInit() error {
	if !hasInit {
		return fmt.Errorf("hackrf: Init was not called")
	}
	return nil
}

func rvToErr(rv C.int) error {
	if rv != 0 {
		errString := C.GoString(C.hackrf_error_name(int32(rv)))
		return fmt.Errorf("hackrf: %s (code: %d)", errString, int32(rv))
	}
	return nil
}

// Board represents the type of HackRf hardware.
type Board uint8

var (
	// BoardInvalid indicates the board that relates to the request is invalid.
	BoardInvalid Board = 0xFF

	// BoardJawbreaker represents a Jawbreaker, the beta test hardware platform
	// for the HackRf.
	BoardJawbreaker Board = 1

	// BoardHackRfOne represents the production HackRf One.
	BoardHackRfOne Board = 2
)

// String will return a human readable string representing the hardware.
func (b Board) String() string {
	switch b {
	case BoardJawbreaker:
		return "Jawbreaker"
	case BoardHackRfOne:
		return "HackRf One"
	case BoardInvalid:
		return "invalid"
	default:
		return "unknown"
	}
}

// List will return the sdr.HardwareInfo for all HackRF devices that are
// plugged into this system.
func List() ([]sdr.HardwareInfo, error) {
	if err := checkInit(); err != nil {
		return nil, err
	}

	list := C.hackrf_device_list()
	defer C.hackrf_device_list_free(list)

	count := int(list.devicecount)
	usbBoardIds := (*[1 << 30]C.enum_hackrf_usb_board_id)(unsafe.Pointer(list.usb_board_ids))[:count]
	serials := (*[1 << 30]*C.char)(unsafe.Pointer(list.serial_numbers))[:count]

	ret := []sdr.HardwareInfo{}
	for i := 0; i < count; i++ {
		ret = append(ret, sdr.HardwareInfo{
			Serial:       C.GoString(serials[i]),
			Manufacturer: "Great Scott Gadgets",
			Product:      Board(usbBoardMapping[usbBoardIds[i]]).String(),
		})
	}

	return ret, nil
}

// Init *must* be called before the HackRF can be used. Attempting to use the
// HackRF or calling any HackRF functions may result in error if done before
// invoking Init.
func Init() error {
	err := rvToErr(C.hackrf_init())
	if err == nil {
		hasInit = true
	}
	return err
}

// Exit will clean up after Init, and should be called when the process is
// shutting down, or otherwise done with the HackRF.
func Exit() error {
	return rvToErr(C.hackrf_exit())
}

// Version will return the HackRF library version and release.
func Version() (string, string) {
	return C.GoString(C.hackrf_library_version()), C.GoString(C.hackrf_library_release())
}

// Open will open the first HackRF on the system.
func Open() (*Sdr, error) {
	var dev *C.hackrf_device

	if err := rvToErr(C.hackrf_open(&dev)); err != nil {
		return nil, err
	}

	return &Sdr{
		dev: dev,
	}, nil
}

// Sdr implements the sdr.Sdr interface for the HackRF One.
type Sdr struct {
	dev *C.hackrf_device

	sampleRate      uint
	centerFrequency rf.Hz
	amp             bool
}

// Close implements the sdr.Sdr interface
func (s *Sdr) Close() error {
	return rvToErr(C.hackrf_close(s.dev))
}

// SetCenterFrequency implements the sdr.Sdr interface
func (s *Sdr) SetCenterFrequency(freq rf.Hz) error {
	err := rvToErr(C.hackrf_set_freq(
		s.dev,
		C.uint64_t(freq),
	))
	if err != nil {
		return err
	}
	s.centerFrequency = freq
	return nil
}

// GetCenterFrequency implements the sdr.Sdr interface
func (s *Sdr) GetCenterFrequency() (rf.Hz, error) {
	return s.centerFrequency, nil
}

// SetAutomaticGain implements the sdr.Sdr interface
func (s *Sdr) SetAutomaticGain(bool) error {
	return sdr.ErrNotSupported
}

// SetSampleRate implements the sdr.Sdr interface
func (s *Sdr) SetSampleRate(sampleRate uint) error {
	err := rvToErr(C.hackrf_set_sample_rate(s.dev, C.double(sampleRate)))
	if err != nil {
		return err
	}
	s.sampleRate = sampleRate
	return nil
}

// GetSampleRate implements the sdr.Sdr interface
func (s *Sdr) GetSampleRate() (uint, error) {
	return s.sampleRate, nil
}

// SampleFormat implements the sdr.Sdr interface
func (s *Sdr) SampleFormat() sdr.SampleFormat {
	return sdr.SampleFormatI8
}

// SetPPM implements the sdr.Sdr interface
func (s *Sdr) SetPPM(int) error {
	return sdr.ErrNotSupported
}

// HardwareInfo implements the sdr.Sdr interface
func (s *Sdr) HardwareInfo() sdr.HardwareInfo {
	var (
		partid = C.read_partid_serialno_t{}
		board  C.uint8_t
	)

	C.hackrf_board_partid_serialno_read(s.dev, &partid)
	C.hackrf_board_id_read(s.dev, &board)

	serial := fmt.Sprintf(
		"%08x%08x%08x%08x",
		partid.serial_no[0],
		partid.serial_no[1],
		partid.serial_no[2],
		partid.serial_no[3],
	)

	return sdr.HardwareInfo{
		Serial:       serial,
		Product:      Board(board).String(),
		Manufacturer: "Great Scott Gadgets",
	}
}

// vim: foldmethod=marker

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

// #cgo linux LDFLAGS: -lhackrf
// #cgo pkg-config: libhackrf
//
// #include <libhackrf/hackrf.h>
import "C"

import (
	"fmt"
	"unsafe"

	"hz.tools/sdr"
)

var (
	hasInit bool = false
)

func checkInit() error {
	if !hasInit {
		return fmt.Errorf("hackrf: hackrf.Init was not called!")
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
type Board uint32

var (
	// BoardInvalid indicates the board that relates to the request is invalid.
	BoardInvalid Board = 0xFFFF

	// BoardJawbreaker represents a Jawbreaker, the beta test hardware platform
	// for the HackRf.
	BoardJawbreaker Board = 0x604B

	// BoardHackRfOne represents the production HackRf One.
	BoardHackRfOne Board = 0x6089
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
			Product:      Board(usbBoardIds[i]).String(),
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

// vim: foldmethod=marker

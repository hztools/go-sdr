// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2022
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

package airspyhf

// #cgo pkg-config: libairspyhf
//
// #include <airspyhf.h>
import "C"

import (
	"bytes"
	"fmt"
	"unsafe"

	"hz.tools/sdr"
	"hz.tools/sdr/debug"
)

// LibraryVersion represents the version of the airspy library that's been
// linked against.
type LibraryVersion struct {
	MajorVersion uint32
	MinorVersion uint32
	Revision     uint32
}

// String will return the LibraryVersion as a semver style dotted version number.
func (lv LibraryVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", lv.MajorVersion, lv.MinorVersion, lv.Revision)
}

// GetLibraryVersion will return the version of the libairspyhf library as
// reported by the C library / airspy bindings.
func GetLibraryVersion() LibraryVersion {
	v := LibraryVersion{}
	C.airspyhf_lib_version((*C.airspyhf_lib_version_t)(unsafe.Pointer(&v)))
	return v
}

// ListSerials enumerates
func ListSerials() []uint64 {
	var (
		ndev       = int(C.airspyhf_list_devices(nil, 0))
		serialSize = 8
		serials    = make([]uint64, ndev*serialSize)
	)
	if ndev == 0 {
		return serials
	}

	ndev = int(C.airspyhf_list_devices(
		(*C.uint64_t)(unsafe.Pointer(&serials[0])),
		C.int(ndev),
	))
	serials = serials[:ndev]
	return serials
}

func OpenBySerial(sn uint64) (*Sdr, error) {
	return open(&sn)
}

func Open() (*Sdr, error) {
	return open(nil)
}

func open(sn *uint64) (*Sdr, error) {
	var (
		dev *C.airspyhf_device_t
		id  C.airspyhf_read_partid_serialno_t
	)

	if sn == nil {
		if C.airspyhf_open(&dev) != C.AIRSPYHF_SUCCESS {
			return nil, fmt.Errorf("airspyhf: Can not open airspy")
		}
	} else {
		if C.airspyhf_open_sn(&dev, C.uint64_t(*sn)) != C.AIRSPYHF_SUCCESS {
			return nil, fmt.Errorf("airspyhf: Can not open airspy sn=%x", *sn)
		}
	}

	if C.airspyhf_board_partid_serialno_read(dev, &id) != C.AIRSPYHF_SUCCESS {
		return nil, fmt.Errorf("airspyhf: Can't real PartID / Serial Number")
	}
	var product string
	switch id.part_id {
	case C.AIRSPYHF_BOARD_ID_UNKNOWN_AIRSPYHF:
		product = "Unknown Airspy HF (0x00)"
	case C.AIRSPYHF_BOARD_ID_AIRSPYHF_REV_A:
		product = "Airspy HF Rev1 (0x01)"
	case C.AIRSPYHF_BOARD_ID_INVALID:
		product = "Invalid (0xFF)"
	default:
		product = fmt.Sprintf("Not Supported (Unknown ID: 0x%.2X)", id.part_id)
	}

	hwInfo := sdr.HardwareInfo{
		Manufacturer: "Airspy",
		Product:      product,
		Serial:       fmt.Sprintf("%x", sn),
	}

	return &Sdr{
		handle: dev,
		info:   hwInfo,
	}, nil
}

func (s *Sdr) Version() (string, error) {
	var out [255]byte
	if C.airspyhf_version_string_read(
		s.handle,
		(*C.char)(unsafe.Pointer(&out[0])),
		C.uint8_t(len(out)),
	) != C.AIRSPYHF_SUCCESS {
		return "", fmt.Errorf("airspyhf.Sdr.Version: failed to get Version string")
	}
	return string(out[:bytes.Index(out[:], []byte{0x00})]), nil
}

type Sdr struct {
	handle *C.airspyhf_device_t
	info   sdr.HardwareInfo
}

func (s *Sdr) Close() error {
	if C.airspyhf_close(s.handle) != C.AIRSPYHF_SUCCESS {
		return fmt.Errorf("airspy.Sdr.Close: Failed to close handle")
	}
	return nil
}

func (s *Sdr) GetSampleRates() ([]uint, error) {
	var nsr C.uint32_t
	if C.airspyhf_get_samplerates(s.handle, &nsr, 0) != C.AIRSPYHF_SUCCESS {
		return nil, fmt.Errorf("airspy.Sdr.GetSampleRates: Can't enumerate number of SampleRates")
	}

	srs := make([]uint32, int(nsr))
	if C.airspyhf_get_samplerates(s.handle, (*C.uint32_t)(unsafe.Pointer(&srs[0])), nsr) != C.AIRSPYHF_SUCCESS {
		return nil, fmt.Errorf("airspy.Sdr.GetSampleRates: Can't enumerate SampleRates")
	}

	ret := make([]uint, len(srs))
	for i := range srs {
		ret[i] = uint(srs[i])
	}

	return ret, nil
}

func (s *Sdr) HardwareInfo() sdr.HardwareInfo {
	return s.info
}

func init() {
	debug.RegisterRadioDriver("hz.tools/sdr/airspyhf.Sdr")
}

// vim: foldmethod=marker

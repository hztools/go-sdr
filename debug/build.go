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

package debug

import (
	"encoding/binary"

	"hz.tools/sdr"
	"hz.tools/sdr/internal"
	"hz.tools/sdr/internal/simd"
)

// SIMDInfo contains information about the status of the SIMD runtime.
type SIMDInfo struct {
	// Enabled is True if using SIMD ASM instructions in the backend, False
	// if using the pure-go implementation.
	Enabled bool

	// Backends is a list of the SIMD backends in use.
	Backends []string
}

// BuildInfo contains information about the compiled support inside this
// library.
type BuildInfo struct {
	// SampleFormats will return all known sdr.SampleFormats understood
	// by this compiled version of hz.tools/sdr
	SampleFormats []sdr.SampleFormat

	// SIMD will return the compile-time SIMD support.
	SIMD SIMDInfo

	// HostEndianness will return the detected host ByteOrder.
	HostEndianness binary.ByteOrder
}

// ReadBuildInfo will return information about the internals of the SDR package,
// including implementation details.
func ReadBuildInfo() BuildInfo {
	return BuildInfo{
		SampleFormats: []sdr.SampleFormat{
			sdr.SampleFormatC64,
			sdr.SampleFormatI16,
			sdr.SampleFormatU8,
			sdr.SampleFormatI8,
		},
		SIMD: SIMDInfo{
			Backends: simd.Backends,
			Enabled:  simd.Enabled,
		},
		HostEndianness: internal.NativeEndian,
	}
}

// vim: foldmethod=marker

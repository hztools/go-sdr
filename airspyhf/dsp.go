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
	"fmt"
)

// SetDSP will enable or disable the following features:
//   - IQ correction
//   - IF shift
//   - Fine Tuning
func (s *Sdr) SetDSP(state bool) error {
	var v C.uint8_t
	if state {
		v = 1
	}
	if C.airspyhf_set_lib_dsp(s.handle, v) != C.AIRSPYHF_SUCCESS {
		return fmt.Errorf("airspyhf.Sdr.SetDSP: failed to set DSP")
	}
	return nil
}

// ConfigureIQBalancer will modify the configuration of the initialized airspy DSP IQ
// balancer.
func (s *Sdr) ConfigureIQBalancer(
	buffersToSkip int,
	fftIntegration int,
	fftOverlap int,
	correlationIntegration int,
) error {
	if C.airspyhf_iq_balancer_configure(
		s.handle,
		C.int(buffersToSkip),
		C.int(fftIntegration),
		C.int(fftOverlap),
		C.int(correlationIntegration),
	) != C.AIRSPYHF_SUCCESS {
		return fmt.Errorf("airspyhf.Sdr.ConfigureIQBalancer: failed to configure the IQ balancer")
	}
	return nil
}

// SetOptimalIQCorrectionPoint will modify the IQ correction point in the
// underlying DSP IQ balancer.
func (s *Sdr) SetOptimalIQCorrectionPoint(w float32) error {
	if C.airspyhf_set_optimal_iq_correction_point(s.handle, C.float(w)) != C.AIRSPYHF_SUCCESS {
		return fmt.Errorf("airspyhf.Sdr.SetOptimalIQCorrectionPoint: failed to set IQ correction point")
	}
	return nil
}

// vim: foldmethod=marker

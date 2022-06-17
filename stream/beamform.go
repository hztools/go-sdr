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

package stream

import (
	"fmt"

	"hz.tools/sdr"
)

// Beamform will combine a set of sdr.Readers into a single sdr.Reader using
// the provided phase angles to stear the beam.
type Beamform struct {
	sdr.Reader

	readers sdr.Readers
	config  BeamformConfig
}

// SetPhaseAngles will set the phase angle to shift every stream by.
func (b *Beamform) SetPhaseAngles(angles []complex64) error {
	if len(angles) != len(b.readers) {
		return fmt.Errorf("Beamform.SetPhaseAngles: angles must match the reader length")
	}
	for i, reader := range b.readers {
		reader.(*multiplyReader).SetMultiplier(angles[i])
	}
	return nil
}

// BeamformConfig contains configuration for the combined samples.
type BeamformConfig struct {
	Angles []complex64
}

// ReadBeamform will create a new sdr.Reader from a series of coherent
// sdr.Readers using the provided phase angles.
func ReadBeamform(rs sdr.Readers, cfg BeamformConfig) (*Beamform, error) {
	multReaders := make(sdr.Readers, len(rs))
	for i := range rs {
		reader, err := ConvertReader(rs[i], sdr.SampleFormatC64)
		if err != nil {
			return nil, err
		}
		multReaders[i], err = Multiply(reader, 1)
		if err != nil {
			return nil, err
		}
	}
	addReader, err := Add(multReaders...)
	if err != nil {
		return nil, err
	}
	b := &Beamform{
		Reader:  addReader,
		readers: multReaders,
		config:  cfg,
	}
	b.SetPhaseAngles(cfg.Angles)
	return b, nil
}

// vim: foldmethod=marker

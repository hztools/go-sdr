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

package mock

import (
	"fmt"

	"hz.tools/rf"
	"hz.tools/sdr"
)

// New will create a new mock sdr.
func New(cfg Config) sdr.Transceiver {
	return &mockSdr{
		config:    &cfg,
		gainState: make(map[string]float32),
	}
}

type mockSdr struct {
	config    *Config
	gainState map[string]float32
}

// Config is the set of default values, and optional features of the MockSDR.
type Config struct {
	// CenterFrequency is the initial CenterFrequency in Hz.
	CenterFrequency rf.Hz

	// SampleRate is the default initial SampleRate.
	SampleRate uint

	// SampleFormat is the format that this SDR speaks, as well as both the
	// Rx and Tx. If this is not set, or this does not match the Rx or Tx
	// Reader or Writer, StartRx and StartTx will return sdr.ErrNotSupported
	SampleFormat sdr.SampleFormat

	// Rx, if not nil, will be used as the data returned by `StartRx`. If this
	// is nil, StartRx will return sdr.ErrNotSupported.
	Rx func(sdr.Transceiver) (sdr.ReadCloser, error)

	// Tx, if not nil, will be used as the data returned by `StartTx`. If this
	// is nil, StartTx will return sdr.ErrNotSupported
	Tx func(sdr.Transceiver) (sdr.WriteCloser, error)

	// GainStages if not nil, will be used as the gain stages supported by the
	// MockSDR. If nil, No gain stages will be returned or settable.
	GainStages sdr.GainStages
}

func (m *mockSdr) HardwareInfo() sdr.HardwareInfo {
	return sdr.HardwareInfo{
		Manufacturer: "hz.tools",
		Product:      "mocksdr",
		Serial:       "",
	}
}

// Close implements the sdr.Sdr interface.
func (m *mockSdr) Close() error {
	return nil
}

// SetCenterFrequency implements the sdr.Sdr interface.
func (m *mockSdr) SetCenterFrequency(r rf.Hz) error {
	m.config.CenterFrequency = r
	return nil
}

// GetCenterFrequency implements the sdr.Sdr interface.
func (m *mockSdr) GetCenterFrequency() (rf.Hz, error) {
	return m.config.CenterFrequency, nil
}

// SetAutomaticGain implements the sdr.Sdr interface.
func (m *mockSdr) SetAutomaticGain(bool) error {
	return sdr.ErrNotSupported
}

// GetGainStages implements the sdr.Sdr interface.
func (m *mockSdr) GetGainStages() (sdr.GainStages, error) {
	if m.config.GainStages == nil {
		return nil, sdr.ErrNotSupported
	}
	return m.config.GainStages, nil
}

// GetGain implements the sdr.Sdr interface.
func (m *mockSdr) GetGain(gs sdr.GainStage) (float32, error) {
	if m.config.GainStages == nil {
		return 0, sdr.ErrNotSupported
	}
	val, ok := m.gainState[gs.String()]
	if ok {
		return val, nil
	}
	return 0, fmt.Errorf("mock: gain not set")
}

// SetGain implements the sdr.Sdr interface.
func (m *mockSdr) SetGain(gs sdr.GainStage, gain float32) error {
	if m.config.GainStages == nil {
		return sdr.ErrNotSupported
	}
	m.gainState[gs.String()] = gain
	return nil
}

// SetSampleRate implements the sdr.Sdr interface.
func (m *mockSdr) SetSampleRate(sps uint) error {
	m.config.SampleRate = sps
	return nil
}

// GetSampleRate implements the sdr.Sdr interface.
func (m *mockSdr) GetSampleRate() (uint, error) {
	return m.config.SampleRate, nil
}

// GetSamplesPerWindow implements the sdr.Sdr interface.
func (m *mockSdr) GetSamplesPerWindow() (uint, error) {
	return m.config.SampleRate, nil
}

// SampleFormat implements the sdr.Sdr interface.
func (m *mockSdr) SampleFormat() sdr.SampleFormat {
	return m.config.SampleFormat
}

// StartRx implements the sdr.Sdr interface.
func (m *mockSdr) StartRx() (sdr.ReadCloser, error) {
	if m.config.Rx == nil {
		return nil, sdr.ErrNotSupported
	}

	rx, err := m.config.Rx(m)
	if err != nil {
		return nil, err
	}

	if rx.SampleFormat() != m.config.SampleFormat {
		return nil, sdr.ErrNotSupported
	}
	return rx, nil
}

func (m *mockSdr) StartTx() (sdr.WriteCloser, error) {
	if m.config.Tx == nil {
		return nil, sdr.ErrNotSupported
	}

	tx, err := m.config.Tx(m)
	if err != nil {
		return nil, err
	}

	if tx.SampleFormat() != m.config.SampleFormat {
		return nil, sdr.ErrNotSupported
	}
	return tx, nil
}

// ThisRx will create a "StartRx" function for a mock.Sdr that simply returns
// the provided ReadCloser. Multiple calls will return the same ReadCloser,
// which means the Close method is likly to be called more than once. Caller
// beware!
func ThisRx(rx sdr.ReadCloser) func(sdr.Transceiver) (sdr.ReadCloser, error) {
	return func(sdr.Transceiver) (sdr.ReadCloser, error) {
		return rx, nil
	}
}

// ThisTx will create a "StartTx" function for a mock.Sdr that simply returns
// the provided WriteCloser. Multiple calls will return the same WriteCloser,
// which means the Close method is likly to be called more than once. Caller
// beware!
func ThisTx(tx sdr.WriteCloser) func(sdr.Transceiver) (sdr.WriteCloser, error) {
	return func(sdr.Transceiver) (sdr.WriteCloser, error) {
		return tx, nil
	}
}

// vim: foldmethod=marker

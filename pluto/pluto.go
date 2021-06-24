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

package pluto

import (
	"fmt"

	"hz.tools/rf"
	"hz.tools/sdr"
	"hz.tools/sdr/debug"
	"hz.tools/sdr/pluto/iio"
)

func init() {
	debug.RegisterRadioDriver("hz.tools/sdr/pluto.Sdr")
}

var (
	// plutoPhyName is the name of the transceiver itself, used for control
	// over things like samples per second or frequency.
	plutoPhyName string = "ad9361-phy"

	// plutoRxName is the name of the RX ADC, to read samples from the rx
	// antenna.
	plutoRxName string = "cf-ad9361-lpc"

	// plutoTxName is the name of the TX DAC, to write samples to the
	// tx side of the house.
	plutoTxName string = "cf-ad9361-dds-core-lpc"
)

// Sdr is an interface to the underlying PlutoSDR endpoint. This will allow
// the user to interact with the Pluto as any other hz.tools/sdr.Sdr. This
// implements both the Receiver and Transmitter (Transceiver) interface.
type Sdr struct {
	endpoint    string
	ictx        *iio.Context
	phy         *iio.Device
	voltage0Rx  *iio.Channel
	voltage0Tx  *iio.Channel
	altVoltage0 *iio.Channel
	altVoltage1 *iio.Channel

	rx *rx
	tx *tx

	txWindowSize     int
	rxWindowSize     int
	samplesPerSecond uint
}

// Open will establish a connection to a PlutoSDR, and return a handle to
// interact with that device. The endpoint string is the URI that would be
// passed to the iio* tools, such as ip:192.168.2.1, or ip:pluto3.hz.tools
func Open(endpoint string) (*Sdr, error) {
	var (
		rxWindowSize int = 4096 * 32
		txWindowSize int = 4096 * 32 * 10
	)

	ictx, err := iio.Open(endpoint)
	if err != nil {
		return nil, err
	}

	phy, err := ictx.FindDevice(plutoPhyName)
	if err != nil {
		return nil, err
	}

	altVoltage0, err := phy.FindChannel("altvoltage0", iio.ChannelDirectionWrite)
	if err != nil {
		return nil, err
	}

	altVoltage1, err := phy.FindChannel("altvoltage1", iio.ChannelDirectionWrite)
	if err != nil {
		return nil, err
	}

	voltage0Rx, err := phy.FindChannel("voltage0", iio.ChannelDirectionRead)
	if err != nil {
		return nil, err
	}

	voltage0Tx, err := phy.FindChannel("voltage0", iio.ChannelDirectionWrite)
	if err != nil {
		return nil, err
	}

	rx, err := openRx(ictx, rxWindowSize)
	if err != nil {
		return nil, err
	}

	tx, err := openTx(ictx, txWindowSize)
	if err != nil {
		return nil, err
	}

	return &Sdr{
		endpoint: endpoint,

		ictx:        ictx,
		phy:         phy,
		altVoltage0: altVoltage0,
		altVoltage1: altVoltage1,
		voltage0Rx:  voltage0Rx,
		voltage0Tx:  voltage0Tx,

		rxWindowSize: rxWindowSize,
		txWindowSize: txWindowSize,

		rx: rx,
		tx: tx,
	}, nil
}

// SetLoopback will set BIST Loopback to send TX data to the RX port, to do
// things like determine the phase offset between RX and TX.
func (s *Sdr) SetLoopback(b bool) error {
	// s.phy
	// loopback

	// 0  Disable
	// 1  Digital TX → Digital RX
	// 2  RF RX → RF TX

	var v int64 = 0
	if b {
		v = 1
	}

	return s.phy.WriteDebugInt64("loopback", v)
}

// Close implements the sdr.Sdr interface.
func (s *Sdr) Close() error {
	return nil
}

// HardwareInfo implements the sdr.Sdr interface
func (s *Sdr) HardwareInfo() sdr.HardwareInfo {
	info := sdr.HardwareInfo{
		Manufacturer: "Analog Devices",
	}

	if model := s.ictx.Attr("hw_model"); model != nil {
		info.Product = *model
	}

	if serial := s.ictx.Attr("hw_serial"); serial != nil {
		info.Serial = *serial
	}

	return info
}

// SetCenterFrequency implements the sdr.Sdr interface.
func (s *Sdr) SetCenterFrequency(r rf.Hz) error {
	if err := s.altVoltage0.WriteInt64("frequency", int64(r)); err != nil {
		return err
	}
	if err := s.altVoltage1.WriteInt64("frequency", int64(r)); err != nil {
		return err
	}
	return nil
}

// GetCenterFrequency implements the sdr.Sdr interface.
func (s *Sdr) GetCenterFrequency() (rf.Hz, error) {
	rxFreq, err := s.altVoltage0.ReadInt64("frequency")
	if err != nil {
		return rf.Hz(0), err
	}

	txFreq, err := s.altVoltage1.ReadInt64("frequency")
	if err != nil {
		return rf.Hz(0), err
	}

	if rxFreq != txFreq {
		return rf.Hz(0), fmt.Errorf("pluto: rx and tx frequencies are different")
	}

	return rf.Hz(rxFreq), nil
}

// SetSampleRate implements the sdr.Sdr interface.
func (s *Sdr) SetSampleRate(sps uint) error {
	if sps < 2083336 {
		// TODO(paultag): Add in decimation bits.
		return fmt.Errorf("pluto: minimum samples per second is 2083336")
	}

	// TODO(paultag): Add in a log message when going above 5msps because
	// it will result in dropped samples.
	//
	// if sps > 5000000 {
	// }

	// TODO(paultag): The tx and rx should be independently controllable
	// for full duplex devices such as Pluto.
	if err := s.voltage0Rx.WriteInt64("sampling_frequency", int64(sps)); err != nil {
		return err
	}
	if err := s.voltage0Rx.WriteInt64("rf_bandwidth", int64(sps)); err != nil {
		return err
	}
	if err := s.voltage0Tx.WriteInt64("sampling_frequency", int64(sps)); err != nil {
		return err
	}
	if err := s.voltage0Tx.WriteInt64("rf_bandwidth", int64(sps)); err != nil {
		return err
	}

	s.samplesPerSecond = sps

	return nil
}

// GetSampleRate implements the sdr.Sdr interface.
func (s *Sdr) GetSampleRate() (uint, error) {
	return 0, sdr.ErrNotSupported
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

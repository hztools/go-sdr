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

package rtltcp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"

	"hz.tools/rf"
	"hz.tools/sdr"
	"hz.tools/sdr/debug"
	"hz.tools/sdr/rtl"
)

func init() {
	debug.RegisterRadioDriver("hz.tools/sdr/rtltcp.Client")
}

// Client is an rtltcp "SDR" implementation.
//
// This implements the sdr.Sdr interface. This allows any code that uses a
// hz.tools/sdr.Sdr to use an rtl-sdr over the network using the standard
// `rtl_tcp` command.
type Client struct {
	conn       net.Conn
	dongleInfo DongleInfo
	windowSize uint
	reader     sdr.Reader
	sampleRate uint
}

// Close will close the underlying net.Conn.
func (c *Client) Close() error {
	return c.conn.Close()
}

// HardwareInfo implements the sdr.Sdr interface
func (c *Client) HardwareInfo() sdr.HardwareInfo {
	// TODO(paultag): Implement me
	return sdr.HardwareInfo{}
}

func bool2uint32(yn bool) uint32 {
	var argument uint32 = 0
	if yn {
		argument = 1
	}
	return argument
}

// Dial will open a connection to the underlying network connection, and
// create a new rtltcp.Client on top of that connection.
//
// Unlike a regular Sdr, the receive side starts as soon as the connection
// is opened, and if the `StartRx` method is not called and consumed as soon
// as opening the connection, it may lead to a backlog of sample windows, and
// the connection may get broken.
//
// Be sure to start consuming from the reader as soon as the Sdr is opened!
func Dial(network, address string) (*Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}

	var di DongleInfo
	if err := binary.Read(conn, binary.BigEndian, &di); err != nil {
		return nil, err
	}

	if bytes.Compare(di.Magic[:], []byte{'R', 'T', 'L', '0'}) != 0 {
		return nil, fmt.Errorf("rtltcp: magic is not RTL0")
	}

	return &Client{
		windowSize: 16 * 32 * 512,
		dongleInfo: di,
		conn:       conn,
		// TODO(paultag): 0 SPS is wrong here, but we also don't know since
		// we only get the samples per second after it's set by the client.
		//
		// Since we're going to go ahead and wrap this with the readerConn
		// below, we'll squirrel away the sample rate in the parent object
		// and return that dynamically.
		reader: sdr.ByteReader(conn, binary.LittleEndian, 0, sdr.SampleFormatU8),
	}, nil
}

type readerConn struct {
	sdr.Reader

	client *Client
}

func (c readerConn) SampleRate() uint {
	return c.client.sampleRate
}

// Close implements the sdr.ReadCloser interface.
func (c readerConn) Close() error {
	return c.client.Close()
}

// SendCommand will send a rtltcp.Request to the Server.
func (c *Client) SendCommand(req Request) error {
	return binary.Write(c.conn, binary.BigEndian, req)
}

// StartRx implements the sdr.Sdr interface
func (c *Client) StartRx() (sdr.ReadCloser, error) {
	// TODO(paultag): Handle the Context in a much better way.
	return readerConn{
		Reader: sdr.ByteReader(c.conn, binary.LittleEndian, 0, sdr.SampleFormatU8),
		client: c,
	}, nil
}

// Tuner implements the sdr.Sdr interface
func (c *Client) Tuner() rtl.Tuner {
	return c.dongleInfo.Tuner()
}

// GetGain implements the sdr.Sdr interface
func (c *Client) GetGain(gainStage sdr.GainStage) (float32, error) {
	return 0, sdr.ErrNotSupported
}

// SetGain implements the sdr.Sdr interface
func (c *Client) SetGain(gainStage sdr.GainStage, gain float32) error {
	switch {
	case gainStage.Type().Is(sdr.GainStageTypeBB):
		return c.SendCommand(Request{
			Command:  CommandSetGain,
			Argument: uint32(gain * 10),
		})
	// TODO(paultag): Add in IF support.
	default:
		return sdr.ErrNotSupported
	}
}

// GetGainStages implements the sdr.Sdr interface.
func (c *Client) GetGainStages() (sdr.GainStages, error) {
	return c.Tuner().GetGainStages()
}

// SetCenterFrequency implements the sdr.Sdr interface
func (c *Client) SetCenterFrequency(freq rf.Hz) error {
	return c.SendCommand(Request{
		Command:  CommandSetFreq,
		Argument: uint32(freq),
	})
}

// GetCenterFrequency implements the sdr.Sdr interface
func (c *Client) GetCenterFrequency() (rf.Hz, error) {
	return 0, sdr.ErrNotSupported
}

// SetAutomaticGain implements the sdr.Sdr interface
func (c *Client) SetAutomaticGain(yn bool) error {
	return c.SendCommand(Request{
		Command:  CommandSetAGCMode,
		Argument: bool2uint32(yn),
	})
}

// SetSampleRate implements the sdr.Sdr interface
func (c *Client) SetSampleRate(rate uint) error {
	err := c.SendCommand(Request{
		Command:  CommandSetSampleRate,
		Argument: uint32(rate),
	})
	if err == nil {
		c.sampleRate = rate
	}
	return err
}

// GetSampleRate implements the sdr.Sdr interface
func (c *Client) GetSampleRate() (uint, error) {
	return c.sampleRate, nil
}

// GetSamplesPerWindow implements the sdr.Sdr interface
func (c *Client) GetSamplesPerWindow() (uint, error) {
	return uint(c.windowSize), nil
}

// SampleFormat implements the sdr.Sdr interface
func (c *Client) SampleFormat() sdr.SampleFormat {
	return sdr.SampleFormatU8
}

// SetTestMode implements the sdr.Sdr interface
func (c *Client) SetTestMode(yn bool) error {
	return c.SendCommand(Request{
		Command:  CommandSetTestMode,
		Argument: bool2uint32(yn),
	})
}

// vim: foldmethod=marker

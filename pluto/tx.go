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
	"sync"
	"unsafe"

	"hz.tools/sdr"
	"hz.tools/sdr/pluto/iio"
)

type tx struct {
	txi        *iio.Channel
	txq        *iio.Channel
	dac        *iio.Device
	windowSize int
}

func openTx(ictx *iio.Context, windowSize int) (*tx, error) {
	dds, err := ictx.FindDevice(plutoTxName)
	if err != nil {
		return nil, err
	}

	txi, err := dds.FindChannel("voltage0", iio.ChannelDirectionWrite)
	if err != nil {
		return nil, err
	}

	txq, err := dds.FindChannel("voltage1", iio.ChannelDirectionWrite)
	if err != nil {
		return nil, err
	}

	return &tx{
		txi:        txi,
		txq:        txq,
		dac:        dds,
		windowSize: windowSize,
	}, nil
}

type writeCloser struct {
	wg     *sync.WaitGroup
	writer sdr.PipeWriter
	reader sdr.PipeReader
	sdr    *Sdr
	buf    sdr.SamplesI16
}

func (wc *writeCloser) Write(iq sdr.Samples) (int, error) {
	return wc.writer.Write(iq)
}

func (wc *writeCloser) Close() error {
	wc.reader.Close()
	wc.wg.Wait()
	return nil
}

func (wc *writeCloser) SampleRate() uint {
	return wc.reader.SampleRate()
}

func (wc *writeCloser) SampleFormat() sdr.SampleFormat {
	return wc.reader.SampleFormat()
}

func (wc *writeCloser) powerup() error {
	return wc.sdr.altVoltage1.WriteBool("powerdown", false)
}

func (wc *writeCloser) powerdown() error {
	return wc.sdr.altVoltage1.WriteBool("powerdown", true)
}

func (wc *writeCloser) run() error {
	defer wc.wg.Done()
	tx := wc.sdr.tx

	tx.txi.Enable()
	tx.txq.Enable()
	defer tx.txi.Disable()
	defer tx.txq.Disable()

	buf := wc.buf
	ibuf, err := tx.dac.CreateBuffer(len(buf))
	if err != nil {
		return err
	}
	defer ibuf.Close()
	if err := tx.dac.ClearCheckBuffer(); err != nil {
		return err
	}

	wc.powerup()
	defer wc.powerdown()

	for {
		if err := tx.dac.CheckBuffer(); err != nil {
			return err
		}

		_, err := sdr.ReadFull(wc.reader, buf)
		if err != nil {
			return err
		}

		_, err = ibuf.CopyToBufferFromUnsafe(
			*tx.txi,
			unsafe.Pointer(&buf[0]),
			buf.Size(),
		)
		if err != nil {
			return err
		}

		if _, err := ibuf.Push(); err != nil {
			return err
		}
	}

	return nil
}

// StartTx implements the sdr.Sdr interface.
func (s *Sdr) StartTx() (sdr.WriteCloser, error) {
	pipeReader, pipeWriter := sdr.Pipe(s.samplesPerSecond, sdr.SampleFormatI16)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	wc := &writeCloser{
		wg:     wg,
		writer: pipeWriter,
		reader: pipeReader,
		sdr:    s,
		buf:    make(sdr.SamplesI16, s.txWindowSize),
	}

	go func() {
		if err := wc.run(); err != nil {
			pipeWriter.CloseWithError(err)
		}
	}()

	return wc, nil

}

// vim: foldmethod=marker

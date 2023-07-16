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
	"context"
	"sync"
	"unsafe"

	"hz.tools/sdr"
	"hz.tools/sdr/internal/iio"
)

// CyclicTx will transmit a specific buffer in a loop. If the buffer is to be
// updated after the fact, be sure to invoke 'Push'.
type CyclicTx struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup

	sdr *Sdr

	ibuf *iio.Buffer
	buf  sdr.SamplesI16
}

// Close will cancel the context, and then wait for the cleanup routine to
// close out any resources it's holding on to.
func (ct *CyclicTx) Close() error {
	ct.cancel()
	ct.wg.Wait()
	return nil
}

// Powerup will turn on the transmit power.
func (ct *CyclicTx) Powerup() error {
	return ct.sdr.altVoltage1.WriteBool("powerdown", false)
}

// Powerdown will turn off the transmit power.
func (ct *CyclicTx) Powerdown() error {
	return ct.sdr.altVoltage1.WriteBool("powerdown", true)
}

// Push will update the PlutoSDR with the buffer provided when the CyclicTx
// was created.
func (ct *CyclicTx) Push() error {
	if ct.ibuf != nil {
		ct.ibuf.Close()
	}

	ibuf, err := ct.sdr.tx.dac.CreateCyclicBuffer(ct.buf.Length())
	if err != nil {
		return err
	}

	_, err = ibuf.CopyToBufferFromUnsafe(
		*ct.sdr.tx.txi,
		unsafe.Pointer(&ct.buf[0]),
		ct.buf.Size(),
	)
	if err != nil {
		return err
	}

	if _, err := ibuf.Push(); err != nil {
		return err
	}

	ct.ibuf = ibuf
	return nil
}

// StartCyclicTx allows for the creation of a cyclic transmit buffer, and to
// manage the updating of that buffer.
func (s *Sdr) StartCyclicTx(buf sdr.SamplesI16) (*CyclicTx, error) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)

	tx := s.tx
	tx.txi.Enable()
	tx.txq.Enable()

	ct := &CyclicTx{
		wg:     wg,
		ctx:    ctx,
		cancel: cancel,
		sdr:    s,
		buf:    buf,
	}

	if err := ct.Push(); err != nil {
		tx.txi.Disable()
		tx.txq.Disable()
		return nil, err
	}
	ct.Powerup()

	go func() {
		defer wg.Done()
		defer tx.txi.Disable()
		defer tx.txq.Disable()
		defer ct.ibuf.Close()
		defer ct.Powerdown()
		<-ctx.Done()
	}()

	return ct, nil

}

// vim: foldmethod=marker

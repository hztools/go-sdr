// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2021
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

	"hz.tools/sdr"
)

// Loopback will enable Pluto's Loopback mode, and return a TX and RX stream
// to be used like an sdr.Pipe. This will put your IQ data through the signal
// path and into your RX channel.
//
// Loopback mode is only disabled on context cancel. Please remember to
// cancel your context!
func (p *Sdr) Loopback(ctx context.Context) (sdr.Reader, sdr.Writer, error) {
	if err := p.SetLoopback(true); err != nil {
		return nil, nil, err
	}

	rx, err := p.StartRx()
	if err != nil {
		return nil, nil, err
	}

	tx, err := p.StartTx()
	if err != nil {
		rx.Close()
		return nil, nil, err
	}

	go func() {
		<-ctx.Done()
		tx.Close()
		rx.Close()
		p.SetLoopback(false)
	}()

	return rx, tx, nil
}

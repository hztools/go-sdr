// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2020-2023
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

package adrv9364

import (
	"hz.tools/sdr/debug"
)

func init() {
	debug.RegisterRadioDriver("hz.tools/sdr/adrv9364.Sdr")
}

var (
	// adrv9364PhyName is the name of the transceiver itself, used for control
	// over things like samples per second or frequency.
	adrv9364PhyName = "ad9361-phy"

	// adrv9364RxName is the name of the RX ADC, to read samples from the rx
	// antenna.
	adrv9364RxName = "cf-ad9361-lpc"

	// adrv9364TxName is the name of the TX DAC, to write samples to the
	// tx side of the house.
	adrv9364TxName = "cf-ad9361-dds-core-lpc"
)

// vim: foldmethod=marker

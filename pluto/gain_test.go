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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGainRangeRx(t *testing.T) {
	_, err := rxHardwareGain.Clamp(1000)
	assert.Error(t, err)

	_, err = rxHardwareGain.Clamp(-100)
	assert.Error(t, err)
}

func TestGainRangeTx(t *testing.T) {
	_, err := txHardwareGain.Clamp(1)
	assert.Error(t, err)

	_, err = txHardwareGain.Clamp(-90)
	assert.Error(t, err)
}

func TestGainStepsRx(t *testing.T) {
	gain, err := rxHardwareGain.Clamp(1.25)
	assert.NoError(t, err)
	assert.Equal(t, 1.0, gain)

	gain, err = rxHardwareGain.Clamp(1.55)
	assert.NoError(t, err)
	assert.Equal(t, 2.0, gain)
}

func TestGainStepsTx(t *testing.T) {
	gain, err := txHardwareGain.Clamp(-1.25)
	assert.NoError(t, err)
	assert.Equal(t, -1.25, gain)

	gain, err = txHardwareGain.Clamp(-1.55)
	assert.NoError(t, err)
	assert.Equal(t, -1.5, gain)

	gain, err = txHardwareGain.Clamp(-1.9)
	assert.NoError(t, err)
	assert.Equal(t, -2.0, gain)
}

// vim: foldmethod=marker

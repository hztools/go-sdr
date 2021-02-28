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

package e4k_test

import (
	//	"fmt"
	"testing"

	"hz.tools/sdr/rtl/e4k"

	"github.com/stretchr/testify/assert"
)

func TestGainSetGet(t *testing.T) {
	stages := e4k.Stages{}
	stages.SetGain(1, 10)
	stages.SetGain(6, 10)
	assert.Equal(t, float32(2), stages.GetGain())
}

func TestGainLookups(t *testing.T) {
	for i := 3; i < 55; i++ {
		stages, err := e4k.IFGainStages(uint(i))
		assert.NoError(t, err)
		assert.Equal(t, float32(i), stages.GetGain())
	}
}

// vim: foldmethod=marker

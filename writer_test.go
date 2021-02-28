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

package sdr_test

import (
	"sync"

	"github.com/stretchr/testify/assert"
	"testing"

	"hz.tools/sdr"
)

func TestMultiWriterU8(t *testing.T) {
	wg := sync.WaitGroup{}

	pipeReader1, pipeWriter1 := sdr.Pipe(0, sdr.SampleFormatU8)
	pipeReader2, pipeWriter2 := sdr.Pipe(0, sdr.SampleFormatU8)

	readAll := func(t *testing.T, r sdr.Reader) {
		defer wg.Done()
		buf := make(sdr.SamplesU8, 1024)
		i, err := sdr.ReadFull(r, buf)
		assert.NoError(t, err)
		assert.Equal(t, 1024, i)
	}

	go readAll(t, pipeReader1)
	go readAll(t, pipeReader2)
	wg.Add(2)

	buf := make(sdr.SamplesU8, 1024)
	buf[0][0] = 0xFF
	buf[1][1] = 0xFF

	mw := sdr.MultiWriter(0, sdr.SampleFormatU8, pipeWriter1, pipeWriter2)
	i, err := mw.Write(buf)
	assert.NoError(t, err)
	assert.Equal(t, 1024, i)

	wg.Wait()
}

// vim: foldmethod=marker

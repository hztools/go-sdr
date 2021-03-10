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

package mock_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/rf"
	"hz.tools/sdr"
	"hz.tools/sdr/mock"
)

func TestSetGet(t *testing.T) {
	dev := mock.New(mock.Config{})
	var testFreq rf.Hz = 1090 * rf.MHz

	assert.NoError(t, dev.SetCenterFrequency(testFreq))
	centerFreq, err := dev.GetCenterFrequency()
	assert.NoError(t, err)
	assert.Equal(t, testFreq, centerFreq)

	assert.NoError(t, dev.SetSampleRate(1000))
	sps, err := dev.GetSampleRate()
	assert.NoError(t, err)
	assert.Equal(t, uint(1000), sps)
}

func TestSetRead(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	readCloser, writeCloser := sdr.PipeWithContext(ctx, 0, sdr.SampleFormatI16)

	dev := mock.New(mock.Config{
		Rx:           mock.ThisRx(readCloser),
		SampleFormat: sdr.SampleFormatI16,
	})

	wg := sync.WaitGroup{}
	go func(t *testing.T, readCloser sdr.WriteCloser) {
		defer wg.Done()
		i, err := readCloser.Write(make(sdr.SamplesI16, 10))
		assert.NoError(t, err)
		assert.Equal(t, i, 10)
	}(t, writeCloser)
	wg.Add(1)

	sf := dev.SampleFormat()
	assert.Equal(t, sf, sdr.SampleFormatI16)

	rx, err := dev.StartRx()
	assert.NoError(t, err)

	i, err := rx.Read(make(sdr.SamplesI16, 10))
	assert.NoError(t, err)
	assert.Equal(t, i, 10)

	wg.Wait()
}

func TestSetWrite(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	readCloser, writeCloser := sdr.PipeWithContext(ctx, 0, sdr.SampleFormatI16)

	dev := mock.New(mock.Config{
		Tx:           mock.ThisTx(writeCloser),
		SampleFormat: sdr.SampleFormatI16,
	})

	tx, err := dev.StartTx()
	assert.NoError(t, err)

	wg := sync.WaitGroup{}
	go func(t *testing.T, readCloser sdr.WriteCloser) {
		defer wg.Done()
		i, err := tx.Write(make(sdr.SamplesI16, 10))
		assert.NoError(t, err)
		assert.Equal(t, i, 10)
	}(t, writeCloser)
	wg.Add(1)

	sf := dev.SampleFormat()
	assert.Equal(t, sdr.SampleFormatI16, sf)

	i, err := readCloser.Read(make(sdr.SamplesI16, 10))
	assert.NoError(t, err)
	assert.Equal(t, i, 10)

	wg.Wait()
}

// vim: foldmethod=marker

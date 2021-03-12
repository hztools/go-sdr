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

package sdr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/sdr"
	"hz.tools/sdr/mock"
)

var (
	testGainStageRecv = testGainStage{
		Rng: [2]float32{1, 2},
		Typ: sdr.GainStageTypeRecieve,
		Str: "Recv",
	}

	testGainStageTran = testGainStage{
		Rng: [2]float32{3, 4},
		Typ: sdr.GainStageTypeTransmit,
		Str: "Tran",
	}

	testGainStageSmit = testGainStage{
		Rng: [2]float32{5, 6},
		Typ: sdr.GainStageTypeTransmit,
		Str: "Smit",
	}
)

func TestGainStageString(t *testing.T) {
	for s, gst := range map[string]sdr.GainStageType{
		"*X":  sdr.GainStageTypeRecieve | sdr.GainStageTypeTransmit,
		"RX":  sdr.GainStageTypeRecieve,
		"TX":  sdr.GainStageTypeTransmit,
		"FE":  sdr.GainStageTypeFE,
		"IF":  sdr.GainStageTypeIF,
		"BB":  sdr.GainStageTypeBB,
		"AMP": sdr.GainStageTypeAmp,
	} {
		assert.Equal(t, s, gst.String())
	}
}

type testGainStage struct {
	Rng [2]float32
	Typ sdr.GainStageType
	Str string
}

func (tsg testGainStage) Range() [2]float32 {
	return tsg.Rng
}

func (tsg testGainStage) Type() sdr.GainStageType {
	return tsg.Typ
}

func (tsg testGainStage) String() string {
	return tsg.Str
}

func TestGainStage(t *testing.T) {
	rxtx := sdr.GainStageTypeRecieve | sdr.GainStageTypeTransmit
	assert.True(t, rxtx.Is(sdr.GainStageTypeRecieve))
	assert.True(t, rxtx.Is(sdr.GainStageTypeTransmit))
}

func TestGainStages(t *testing.T) {
	gs := sdr.GainStages{
		testGainStageRecv,
		testGainStageTran,
		testGainStageSmit,
	}

	s := gs.First(sdr.GainStageTypeTransmit)
	assert.Equal(t, "Tran", s.String())

	gsm := gs.Map()
	assert.Equal(t, "Tran", gsm["Tran"].String())
}

func TestSetGainStages(t *testing.T) {

	m := mock.New(mock.Config{
		SampleFormat: sdr.SampleFormatU8,
		GainStages: sdr.GainStages{
			testGainStageTran,
			testGainStageSmit,
		},
	})

	assert.NoError(t, sdr.SetGainStages(m, map[string]float32{
		"Tran": 10,
		"Smit": 100,
	}))

	gain, err := m.GetGain(testGainStageTran)
	assert.NoError(t, err)
	assert.Equal(t, float32(10), gain)

	gain, err = m.GetGain(testGainStageSmit)
	assert.NoError(t, err)
	assert.Equal(t, float32(100), gain)
}

// vim: foldmethod=marker

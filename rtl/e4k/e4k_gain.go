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

package e4k

import (
	"fmt"
)

// Stages represents a specific state of all 6 e4k IF gain stages. We treat
// them as a single unit here since we can't set the IF gain to sensible values
// without setting all 6.
type Stages [6]int

// GetGain will return the total Gain represented by the state of IF Gains.
func (s *Stages) GetGain() float32 {
	var gain float32
	for _, stage := range s {
		gain += (float32(stage) * 0.1)
	}
	return gain
}

// SetGain will set the Gain to a specific value for a given Stage. There is
// no check to see if it is a valid value for that stage, so care should be
// taken to be sure the Stages values are sensible.
func (s *Stages) SetGain(stage uint, gain int) {
	s[stage-1] = gain
}

var (
	// sensitivity mode
	senIFGains = []Stages{
		Stages{-30, 00, 00, 00, 30, 30},  // 6
		Stages{-30, 00, 00, 10, 30, 30},  // 7
		Stages{-30, 00, 00, 20, 30, 30},  // 8
		Stages{-30, 30, 00, 00, 30, 30},  // 9
		Stages{-30, 30, 00, 10, 30, 30},  // 10
		Stages{-30, 30, 00, 20, 30, 30},  // 11
		Stages{-30, 60, 00, 00, 30, 30},  // 12
		Stages{-30, 60, 00, 10, 30, 30},  // 13
		Stages{-30, 60, 00, 20, 30, 30},  // 14
		Stages{60, 00, 00, 00, 30, 30},   // 15
		Stages{60, 00, 00, 10, 30, 30},   // 16
		Stages{60, 00, 00, 20, 30, 30},   // 17
		Stages{60, 30, 00, 00, 30, 30},   // 18
		Stages{60, 30, 00, 10, 30, 30},   // 19
		Stages{60, 30, 00, 20, 30, 30},   // 20
		Stages{60, 60, 00, 00, 30, 30},   // 21
		Stages{60, 60, 00, 10, 30, 30},   // 22
		Stages{60, 60, 00, 20, 30, 30},   // 23
		Stages{60, 90, 00, 00, 30, 30},   // 24
		Stages{60, 90, 00, 10, 30, 30},   // 25
		Stages{60, 90, 00, 20, 30, 30},   // 26
		Stages{60, 90, 30, 00, 30, 30},   // 27
		Stages{60, 90, 30, 10, 30, 30},   // 28
		Stages{60, 90, 30, 20, 30, 30},   // 29
		Stages{60, 90, 60, 00, 30, 30},   // 30
		Stages{60, 90, 60, 10, 30, 30},   // 31
		Stages{60, 90, 60, 20, 30, 30},   // 32
		Stages{60, 90, 90, 00, 30, 30},   // 33
		Stages{60, 90, 90, 10, 30, 30},   // 34
		Stages{60, 90, 90, 20, 30, 30},   // 35
		Stages{60, 90, 90, 00, 60, 30},   // 36
		Stages{60, 90, 90, 10, 60, 30},   // 37
		Stages{60, 90, 90, 20, 60, 30},   // 38
		Stages{60, 90, 90, 00, 90, 30},   // 39
		Stages{60, 90, 90, 10, 90, 30},   // 40
		Stages{60, 90, 90, 20, 90, 30},   // 41
		Stages{60, 90, 90, 00, 120, 30},  // 42
		Stages{60, 90, 90, 10, 120, 30},  // 43
		Stages{60, 90, 90, 20, 120, 30},  // 44
		Stages{60, 90, 90, 00, 150, 30},  // 45
		Stages{60, 90, 90, 10, 150, 30},  // 46
		Stages{60, 90, 90, 20, 150, 30},  // 47
		Stages{60, 90, 90, 00, 150, 60},  // 48
		Stages{60, 90, 90, 10, 150, 60},  // 49
		Stages{60, 90, 90, 20, 150, 60},  // 50
		Stages{60, 90, 90, 00, 150, 90},  // 51
		Stages{60, 90, 90, 10, 150, 90},  // 52
		Stages{60, 90, 90, 20, 150, 90},  // 53
		Stages{60, 90, 90, 00, 150, 120}, // 54
		Stages{60, 90, 90, 10, 150, 120}, // 55
		Stages{60, 90, 90, 20, 150, 120}, // 56
		Stages{60, 90, 90, 00, 150, 150}, // 57
		Stages{60, 90, 90, 10, 150, 150}, // 58
		Stages{60, 90, 90, 20, 150, 150}, // 59
		Stages{60, 90, 90, 30, 150, 150}, // 60
	}

	// linerarity mode
	linIFGains = []Stages{
		Stages{-30, 00, 00, 00, 30, 30},   // 6
		Stages{-30, 00, 00, 10, 30, 30},   // 7
		Stages{-30, 00, 00, 20, 30, 30},   // 8
		Stages{-30, 00, 00, 00, 30, 60},   // 9
		Stages{-30, 00, 00, 10, 30, 60},   // 10
		Stages{-30, 00, 00, 20, 30, 60},   // 11
		Stages{-30, 00, 00, 00, 30, 90},   // 12
		Stages{-30, 00, 00, 10, 30, 90},   // 13
		Stages{-30, 00, 00, 20, 30, 90},   // 14
		Stages{-30, 00, 00, 00, 30, 120},  // 15
		Stages{-30, 00, 00, 10, 30, 120},  // 16
		Stages{-30, 00, 00, 20, 30, 120},  // 17
		Stages{-30, 00, 00, 00, 30, 150},  // 18
		Stages{-30, 00, 00, 10, 30, 150},  // 19
		Stages{-30, 00, 00, 20, 30, 150},  // 20
		Stages{-30, 00, 00, 00, 60, 150},  // 21
		Stages{-30, 00, 00, 10, 60, 150},  // 22
		Stages{-30, 00, 00, 20, 60, 150},  // 23
		Stages{-30, 00, 00, 00, 90, 150},  // 24
		Stages{-30, 00, 00, 10, 90, 150},  // 25
		Stages{-30, 00, 00, 20, 90, 150},  // 26
		Stages{-30, 00, 00, 00, 120, 150}, // 27
		Stages{-30, 00, 00, 10, 120, 150}, // 28
		Stages{-30, 00, 00, 20, 120, 150}, // 29
		Stages{-30, 00, 00, 00, 150, 150}, // 30
		Stages{-30, 00, 00, 10, 150, 150}, // 31
		Stages{-30, 00, 00, 20, 150, 150}, // 32
		Stages{-30, 00, 30, 00, 150, 150}, // 33
		Stages{-30, 00, 30, 10, 150, 150}, // 34
		Stages{-30, 00, 30, 20, 150, 150}, // 35
		Stages{-30, 00, 60, 00, 150, 150}, // 36
		Stages{-30, 00, 60, 10, 150, 150}, // 37
		Stages{-30, 00, 60, 20, 150, 150}, // 38
		Stages{-30, 00, 90, 00, 150, 150}, // 39
		Stages{-30, 00, 90, 10, 150, 150}, // 40
		Stages{-30, 00, 90, 20, 150, 150}, // 41
		Stages{-30, 30, 90, 00, 150, 150}, // 42
		Stages{-30, 30, 90, 10, 150, 150}, // 43
		Stages{-30, 30, 90, 20, 150, 150}, // 44
		Stages{-30, 60, 90, 00, 150, 150}, // 45
		Stages{-30, 60, 90, 10, 150, 150}, // 46
		Stages{-30, 60, 90, 20, 150, 150}, // 47
		Stages{60, 00, 90, 00, 150, 150},  // 48
		Stages{60, 00, 90, 10, 150, 150},  // 49
		Stages{60, 00, 90, 20, 150, 150},  // 50
		Stages{60, 30, 90, 00, 150, 150},  // 51
		Stages{60, 30, 90, 10, 150, 150},  // 52
		Stages{60, 30, 90, 20, 150, 150},  // 53
		Stages{60, 60, 90, 00, 150, 150},  // 54
		Stages{60, 60, 90, 10, 150, 150},  // 55
		Stages{60, 60, 90, 20, 150, 150},  // 56
		Stages{60, 90, 90, 00, 150, 150},  // 57
		Stages{60, 90, 90, 10, 150, 150},  // 58
		Stages{60, 90, 90, 20, 150, 150},  // 59
		Stages{60, 90, 90, 30, 150, 150},  // 60
	}
)

// IFGainStages will compute the given Stages for the IF stages.
//
// Specifically, this will use the linerarity gain profile from the e4k spec
// sheet. In the future this may accept a flag to pick the sensitivity table
// instead.
func IFGainStages(gain uint) (Stages, error) {
	if gain < 3 {
		return Stages{}, fmt.Errorf("rtlsdr/e4k: if gain can't go below 3")
	}
	if gain > 56 {
		return Stages{}, fmt.Errorf("rtlsdr/e4k: if gain can't go above 54db")
	}

	gain = gain - 3

	// return senIFGains[gain], nil
	return linIFGains[gain], nil
}

// vim: foldmethod=marker

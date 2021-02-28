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

package internal

import (
	"sync"

	"hz.tools/sdr"
)

// ReadBuffers will do an sdr.ReadFull for each reader and buffer pair.
func ReadBuffers(readers []sdr.Reader, bufs []sdr.SamplesC64) error {
	var err error
	wg := sync.WaitGroup{}
	wg.Add(len(readers))
	for i := range readers {
		go func(i int) {
			defer wg.Done()
			_, ierr := sdr.ReadFull(readers[i], bufs[i])
			if ierr != nil {
				err = ierr
			}
		}(i)
	}
	wg.Wait()
	return err
}

func scaleComplex(el complex64, scale float32) complex64 {
	return complex(
		real(el)/scale,
		imag(el)/scale,
	)
}

// FFTShiftAndScale will do a FFT frequency flip, and also scale any values
// by a float32 real value (both real and imag) which helps if the resulting
// FFT is unscaled.
func FFTShiftAndScale(data []complex64, scale float32) error {
	half := len(data) / 2
	for i := 0; i < half; i++ {
		data[i], data[half+i] = scaleComplex(data[half+i], scale), scaleComplex(data[i], scale)
		// data[i], data[half+i] = scaleComplex(data[i], scale), scaleComplex(data[half+i], scale)
	}
	return nil
}

// vim: foldmethod=marker

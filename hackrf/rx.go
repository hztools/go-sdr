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

package hackrf

// #cgo pkg-config: libhackrf
//
// #include <libhackrf/hackrf.h>
//
// extern int hackrf_rx_callback(hackrf_transfer* transfer);
import "C"

import (
	"log"
	"unsafe"

	"github.com/mattn/go-pointer"

	"hz.tools/sdr"
)

type rxCallbackState struct {
	pipeReader sdr.PipeReader
	pipeWriter sdr.PipeWriter
}

//export hackrfRxCallback
func hackrfRxCallback(transfer *C.hackrf_transfer) int {
	state := pointer.Restore(transfer.rx_ctx).(*rxCallbackState)

	// First, we need to load the incoming bytes from HackRF to a Go
	// []byte type.
	//
	// Let's first compute bounds to avoid doing weirdo stuff later.
	bufSize := int(transfer.valid_length)
	if bufSize%2 != 0 {
		log.Printf("hackrf: bufSize is misaligned")
		bufSize--
	}
	bufIQLength := bufSize / sdr.SampleFormatI8.Size()

	buf := C.GoBytes(unsafe.Pointer(transfer.buffer), C.int(bufSize))

	// TODO(paultag): use a pool?
	//
	// Now let's allocate a new sdr.Samples, copy the bytes from the
	// []byte above directly onto the IQ samples, and write that into the
	// pipe.
	samples := make(sdr.SamplesI8, bufIQLength)

	if copy(sdr.MustUnsafeSamplesAsBytes(samples), buf) != bufSize {
		log.Printf("hackrf: copy() didn't move the whole window over")
		return -1
	}

	i, err := state.pipeWriter.Write(samples)
	if err != nil {
		log.Printf("hackrf: write error %s", err)
		return -1
	}

	if i != bufIQLength {
		log.Printf("hackrf: short write")
		return -1
	}

	return 0
}

// StartRx implements the sdr.Sdr interface.
func (s *Sdr) StartRx() (sdr.ReadCloser, error) {
	pipeReader, pipeWriter := sdr.Pipe(s.sampleRate, sdr.SampleFormatI8)

	state := pointer.Save(&rxCallbackState{
		pipeReader: pipeReader,
		pipeWriter: pipeWriter,
	})

	if err := rvToErr(C.hackrf_start_rx(
		s.dev,
		C.hackrf_sample_block_cb_fn(C.hackrf_rx_callback),
		state,
	)); err != nil {
		return nil, err
	}

	if err := s.SetCenterFrequency(s.centerFrequency); err != nil {
		return nil, err
	}

	if err := s.SetSampleRate(s.sampleRate); err != nil {
		return nil, err
	}

	var closed bool
	return sdr.ReaderWithCloser(pipeReader, func() error {
		if closed {
			return nil
		}
		defer pointer.Unref(state)
		err := rvToErr(C.hackrf_stop_rx(s.dev))
		pipeWriter.Close()
		closed = true
		return err
	}), nil
}

// vim: foldmethod=marker

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

package hackrf

// #cgo pkg-config: libhackrf
//
// #include <libhackrf/hackrf.h>
//
// extern int hackrf_tx_callback(hackrf_transfer* transfer);
import "C"

import (
	"log"
	"sync"
	"unsafe"

	"github.com/mattn/go-pointer"

	"hz.tools/sdr"
	"hz.tools/sdr/yikes"
)

type txCallbackState struct {
	pipeReader sdr.PipeReader
	pipeWriter sdr.PipeWriter
}

//export hackrfTxCallback
func hackrfTxCallback(transfer *C.hackrf_transfer) int {
	state := pointer.Restore(transfer.tx_ctx).(*txCallbackState)

	// First, we need to load the incoming bytes from HackRF to a Go
	// []byte type.
	//
	// Let's first compute bounds to avoid doing weirdo stuff later.
	bufSize := int(transfer.valid_length)
	if bufSize%2 != 0 {
		log.Printf("hackrf: tx: bufSize is misaligned")
		bufSize--
	}
	buf := yikes.GoBytes(uintptr(unsafe.Pointer(transfer.buffer)), bufSize)
	bufIQLength := bufSize / sdr.SampleFormatI8.Size()

	// Now, let's grab some fresh bytes from the ole' pipe
	samples := make(sdr.SamplesI8, bufIQLength)

	n, err := sdr.ReadFull(state.pipeReader, samples)

	// Here we're going to continue even if there's an error so that we ensure
	// we transmit /everything/ we can.
	if n == 0 && err != nil {
		log.Printf("hackrf: tx: failed to ReadFull: %s", err)
		return -1
	}

	if n != samples.Length() {
		for i := n; i < samples.Length(); i++ {
			// The buffer is trash, let's overwrite with zeros.
			samples[i] = [2]int8{0, 0}
		}
	}

	copy(buf, sdr.MustUnsafeSamplesAsBytes(samples))
	return 0
}

// StartTx implements the sdr.Sdr interface.
func (s *Sdr) StartTx() (sdr.WriteCloser, error) {
	if s.amp {
		if err := ampGain(newSteppedGain(
			"Amp",
			sdr.GainStageTypeAmp,
			0, 14, 14,
		)).SetGain(s, 14); err != nil {
			return nil, err
		}
	}

	pipeReader, pipeWriter := sdr.Pipe(s.sampleRate, sdr.SampleFormatI8)

	state := pointer.Save(&txCallbackState{
		pipeReader: pipeReader,
		pipeWriter: pipeWriter,
	})

	if err := rvToErr(C.hackrf_start_tx(
		s.dev,
		C.hackrf_sample_block_cb_fn(C.hackrf_tx_callback),
		state,
	)); err != nil {
		return nil, err
	}

	var (
		lock   = &sync.Mutex{}
		closed bool
	)
	return sdr.WriterWithCloser(pipeWriter, func() error {
		lock.Lock()
		defer lock.Unlock()

		if closed {
			return nil
		}

		defer pointer.Unref(state)

		pipeWriter.Close()
		for C.hackrf_is_streaming(s.dev) == C.HACKRF_TRUE {
			// yield control back? not ideal.
		}

		err := rvToErr(C.hackrf_stop_tx(s.dev))
		closed = true
		return err
	}), nil
}

// vim: foldmethod=marker

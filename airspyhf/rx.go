// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2022
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

package airspyhf

// #cgo pkg-config: libairspyhf
//
// #include <airspyhf.h>
//
// extern int airspyhf_rx_callback(airspyhf_transfer_t*);
import "C"

import (
	"context"
	"fmt"
	"unsafe"

	"github.com/mattn/go-pointer"

	"hz.tools/sdr"
	"hz.tools/sdr/internal/yikes"
)

type callbackContext struct {
	ctx        context.Context
	pipeReader sdr.PipeReader
	pipeWriter sdr.PipeWriter
}

type rx struct {
	sdr.ReadCloser

	ctx    context.Context
	cancel context.CancelFunc
	s      *Sdr
}

func (rx rx) Close() error {
	// sdr stop callbacks
	rx.cancel()
	if C.airspyhf_stop(rx.s.handle) != C.AIRSPYHF_SUCCESS {
		return fmt.Errorf("airspyhf rx.Close(): failed to stop streaming")
	}
	return nil
}

// typedef int (*airspyhf_sample_block_cb_fn) (airspyhf_transfer_t* transfer_fn);

//export airspyhfRxCallback
func airspyhfRxCallback(transfer *C.airspyhf_transfer_t) C.int {
	context := pointer.Restore(transfer.ctx).(*callbackContext)

	// First, check to see if we need to stop.
	if err := context.ctx.Err(); err != nil {
		context.pipeWriter.CloseWithError(err)
		return -1
	}

	iq, err := yikes.Samples(uintptr(unsafe.Pointer(transfer.samples)),
		int(transfer.sample_count), sdr.SampleFormatC64)
	if err != nil {
		context.pipeWriter.CloseWithError(err)
		return -1
	}

	_, err = context.pipeWriter.Write(iq)
	if err != nil {
		context.pipeWriter.CloseWithError(err)
		return -1
	}

	return 0
}

// StartRx implements the sdr.Receiver interface
func (s *Sdr) StartRx() (sdr.ReadCloser, error) {
	// Before we go off and do anything here, let's
	// check to see if we're currently streaming (due
	// to bad cleanup or something), and if so, stop
	// the rx.

	if C.airspyhf_is_streaming(s.handle) == 1 {
		if C.airspyhf_stop(s.handle) != C.AIRSPYHF_SUCCESS {
			return nil, fmt.Errorf("airspyhf.Sdr.StartRx: can't stop existing stream")
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	sps, err := s.GetSampleRate()
	if err != nil {
		return nil, err
	}

	pipeReader, pipeWriter := sdr.PipeWithContext(ctx, sps, sdr.SampleFormatC64)

	cc := &callbackContext{
		ctx:        ctx,
		pipeReader: pipeReader,
		pipeWriter: pipeWriter,
	}
	state := pointer.Save(cc)

	if C.airspyhf_start(
		s.handle,
		C.airspyhf_sample_block_cb_fn(C.airspyhf_rx_callback),
		state,
	) != C.AIRSPYHF_SUCCESS {
		err = fmt.Errorf("airspyhf.Sdr.StartRx: airspy_start failed")
	}

	go func(r *Sdr, state unsafe.Pointer) {
		defer pointer.Unref(state)
		<-ctx.Done()
		pipeReader.Close()
	}(s, state)

	return rx{
		ReadCloser: pipeReader,
		ctx:        ctx,
		cancel:     cancel,
		s:          s,
	}, nil

}

// vim: foldmethod=marker

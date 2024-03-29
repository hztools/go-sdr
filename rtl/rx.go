// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2020-2022
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

package rtl

// #cgo pkg-config: librtlsdr
//
// #include <stdint.h>
// #include <stdlib.h>
//
// #include <rtl-sdr.h>
//
// void rtlsdr_rx_callback(unsigned char *buf, uint32_t len, void *ctx);
import "C"

import (
	"log"
	"unsafe"

	"github.com/mattn/go-pointer"

	"hz.tools/sdr"
)

type callbackContext struct {
	pipeReader sdr.PipeReader
	pipeWriter sdr.PipeWriter
	pool       *sdr.SamplesPool
}

//export rtlsdrRxCallback
func rtlsdrRxCallback(cBuf *C.char, cBufLen C.uint32_t, ptr unsafe.Pointer) {
	context := pointer.Restore(ptr).(*callbackContext)

	buf := C.GoBytes(unsafe.Pointer(cBuf), C.int(cBufLen))
	samples := context.pool.Get().(sdr.SamplesU8)
	defer context.pool.Put(samples)

	n := copy(sdr.MustUnsafeSamplesAsBytes(samples), buf)
	// n here is in bytes; but since we have two bytes for each sample,
	// we need to cut it in half to deal with the number of samples
	// actually written to the buffer.
	samples = samples[:n/2]

	_, err := context.pipeWriter.Write(samples)
	if err != nil {
		// TODO(paultag): Set an error condition and crater the rx path
		// on this one I think.
		log.Println(err)
	}
}

type rx struct {
	sdr.ReadCloser
	rtlSdr Sdr
}

func (rx rx) Close() error {
	if err := rvToErr(C.rtlsdr_cancel_async(rx.rtlSdr.handle)); err != nil {
		log.Printf("Error stopping rx: %s", err)
	}
	return rx.ReadCloser.Close()
}

// StartRx will start to receive IQ samples, ready for consumption from the
// returned ReadCloser.
func (r Sdr) StartRx() (sdr.ReadCloser, error) {
	sps, err := r.GetSampleRate()
	if err != nil {
		return nil, err
	}

	pipeReader, pipeWriter := sdr.Pipe(sps, sdr.SampleFormatU8)

	if err := r.ResetBuffer(); err != nil {
		return nil, err
	}

	var windowSize uint = 16 * 32 * 512
	if r.windowSize != 0 {
		windowSize = r.windowSize
	}

	pool, err := sdr.NewSamplesPool(sdr.SampleFormatU8, int(windowSize/2))
	if err != nil {
		return nil, err
	}

	cc := &callbackContext{
		pipeReader: pipeReader,
		pipeWriter: pipeWriter,
		pool:       pool,
	}

	state := pointer.Save(cc)

	go func(r Sdr, state unsafe.Pointer) {
		defer pointer.Unref(state)
		err := rvToErr(C.rtlsdr_read_async(
			r.handle,
			C.rtlsdr_read_async_cb_t(C.rtlsdr_rx_callback),
			state, 0, C.uint32_t(windowSize),
		))
		pipeReader.CloseWithError(err)
	}(r, state)

	return rx{
		ReadCloser: pipeReader,
		rtlSdr:     r,
	}, nil
}

// vim: foldmethod=marker

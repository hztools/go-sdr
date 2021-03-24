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

package lime

// #cgo pkg-config: LimeSuite
//
// #include <lime/LimeSuite.h>
import "C"

import (
	"log"
	"unsafe"

	"hz.tools/sdr"
)

func goBytesButReally(
	base uintptr,
	size int,
) []byte {
	var b = struct {
		base uintptr
		len  int
		cap  int
	}{base, size, size}
	return *(*[]byte)(unsafe.Pointer(&b))
}

// StartTx implements the sdr.Transmitter interface.
func (s *Sdr) StartTx() (sdr.WriteCloser, error) {
	phasorSize := 2 * 2 // [2]int16
	txBufferSize := 1024 * 32
	txBufferSizeC := txBufferSize * phasorSize

	pipeReader, pipeWriter := sdr.Pipe(s.sampleRate, s.SampleFormat())

	txStream := C.lms_stream_t{}
	txStream.channel = 0
	txStream.fifoSize = C.uint(txBufferSize * 10)
	txStream.throughputVsLatency = 0.5
	txStream.dataFmt = C.LMS_FMT_I16
	txStream.isTx = tx.api()

	txMeta := C.lms_stream_meta_t{}
	txMeta.waitForTimestamp = false
	txMeta.flushPartialPacket = false
	txMeta.timestamp = 0

	var (
		enabled C.bool  = true
		channel C.ulong = 0
	)
	if err := rvToErr(C.LMS_EnableChannel(s.devPtr(), tx.api(), channel, enabled)); err != nil {
		return nil, err
	}

	if err := rvToErr(C.LMS_SetupStream(s.devPtr(), &txStream)); err != nil {
		return nil, err
	}

	if err := rvToErr(C.LMS_StartStream(&txStream)); err != nil {
		return nil, err
	}

	txBufferC := C.malloc(C.ulong(txBufferSizeC))
	txBuffer := make(sdr.SamplesI16, txBufferSize)
	txBufferBytes := sdr.MustUnsafeSamplesAsBytes(txBuffer)
	txBufferCBytes := goBytesButReally(
		uintptr(unsafe.Pointer(txBufferC)),
		txBufferSizeC,
	)

	go func() {
		defer pipeWriter.Close()
		defer C.free(txBufferC)

		for {
			i, err := sdr.ReadFull(pipeReader, txBuffer)
			if err != nil {
				log.Printf("lime: failed to read tx buffer: %s", err)
				return
			}

			n := copy(txBufferCBytes, txBufferBytes)
			if i*phasorSize != n {
				log.Printf("lime: phasors became unaligned, aborting to avoid bad iq")
				return
			}

			v := C.LMS_SendStream(
				&txStream,
				txBufferC,
				C.ulong(i),
				&txMeta,
				1000,
			)
			if int(v) != i {
				log.Printf("lime: incomplete write")
			}

			if v < 0 {
				err := rvToErr(v)
				log.Printf("lime: LMS_SendStream broke with %s", err)
				return
			}
		}
	}()

	return sdr.WriterWithCloser(pipeWriter, func() error {
		defer pipeWriter.Close()
		if err := rvToErr(C.LMS_StopStream(&txStream)); err != nil {
			return err
		}
		return rvToErr(C.LMS_DestroyStream(s.devPtr(), &txStream))

	}), nil
}

// vim: foldmethod=marker

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

// StartRx implements the sdr.Receiver interface.
func (s *Sdr) StartRx() (sdr.ReadCloser, error) {
	phasorSize := 2 * 2 // [2]int16
	rxBufferSize := 1024 * 32
	rxBufferSizeC := rxBufferSize * phasorSize

	pipeReader, pipeWriter := sdr.Pipe(s.sampleRate, s.SampleFormat())

	rxStream := C.lms_stream_t{}
	rxStream.channel = 0
	rxStream.fifoSize = C.uint(rxBufferSize)
	rxStream.throughputVsLatency = 0.5
	rxStream.dataFmt = C.LMS_FMT_I16
	rxStream.isTx = rx.api()

	// rxMeta := C.lms_stream_meta_t{}
	// rxMeta.waitForTimestamp = false
	// rxMeta.flushPartialPacket = false
	// rxMeta.timestamp = 0

	var (
		enabled C.bool  = true
		channel C.ulong = 0
	)
	if err := rvToErr(C.LMS_EnableChannel(s.devPtr(), rx.api(), channel, enabled)); err != nil {
		return nil, err
	}

	if err := rvToErr(C.LMS_SetupStream(s.devPtr(), &rxStream)); err != nil {
		return nil, err
	}

	if err := rvToErr(C.LMS_StartStream(&rxStream)); err != nil {
		return nil, err
	}

	rxBufferC := C.malloc(C.ulong(rxBufferSizeC))
	rxBuffer := make(sdr.SamplesI16, rxBufferSize)
	rxBufferBytes := sdr.MustUnsafeSamplesAsBytes(rxBuffer)

	go func() {
		defer pipeWriter.Close()
		defer C.free(rxBufferC)

		for {
			v := C.LMS_RecvStream(
				&rxStream,
				rxBufferC,
				C.ulong(rxBufferSize),
				nil, // &rxMeta,
				1000,
			)

			if v < 0 {
				err := rvToErr(v)
				log.Printf("LMS_RecvStream: %s", err)
				return
			}

			rxBufferCBytes := C.GoBytes(unsafe.Pointer(rxBufferC), C.int(rxBufferSizeC))
			i := copy(rxBufferBytes, rxBufferCBytes)
			if i == int(v) {
				log.Printf("copy mismatched LMS_RecvStream")
				return
			}

			if i%phasorSize != 0 {
				log.Printf("copy misaligned phasors")
			}

			_, err := pipeWriter.Write(rxBuffer.Slice(0, i/phasorSize))
			if err != nil {
				log.Printf("writer.Write: %s", err)
				return
			}
		}
	}()

	return sdr.ReaderWithCloser(pipeReader, func() error {
		defer pipeWriter.Close()
		if err := rvToErr(C.LMS_StopStream(&rxStream)); err != nil {
			return err
		}
		return rvToErr(C.LMS_DestroyStream(s.devPtr(), &rxStream))

	}), nil
}

// vim: foldmethod=marker

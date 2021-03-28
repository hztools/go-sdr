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
	sampleFormat := s.SampleFormat()
	phasorSize := sampleFormat.Size()

	rxBufferSize := s.options.getBufferSize()
	rxBufferSizeC := rxBufferSize * phasorSize

	pipeReader, pipeWriter := sdr.Pipe(s.sampleRate, sampleFormat)

	rxStream := C.lms_stream_t{}
	rxStream.channel = C.uint(s.options.getChannel())
	rxStream.fifoSize = C.uint(rxBufferSize)
	rxStream.throughputVsLatency = 0.5
	rxStream.isTx = rx.api()

	switch sampleFormat {
	case sdr.SampleFormatI16:
		rxStream.dataFmt = C.LMS_FMT_I16
	case sdr.SampleFormatC64:
		rxStream.dataFmt = C.LMS_FMT_F32
	default:
		return nil, sdr.ErrSampleFormatUnknown
	}

	// rxMeta := C.lms_stream_meta_t{}
	// rxMeta.waitForTimestamp = false
	// rxMeta.flushPartialPacket = false
	// rxMeta.timestamp = 0

	var (
		enabled C.bool  = true
		channel C.ulong = C.ulong(s.options.getChannel())
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
	rxBuffer, err := sdr.MakeSamples(sampleFormat, rxBufferSize)
	if err != nil {
		return nil, err
	}
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
				log.Printf("lime: LMS_RecvStream broke with %s", err)
				return
			}

			rxBufferCBytes := C.GoBytes(unsafe.Pointer(rxBufferC), C.int(rxBufferSizeC))
			i := copy(rxBufferBytes, rxBufferCBytes)
			if int(v)*phasorSize != i {
				log.Printf("lime: phasors became unaligned, aborting to avoid bad iq")
				return
			}

			if i%phasorSize != 0 {
				log.Printf("lime: somhow we don't have things aligned to [2]int16 bounds")
				return
			}

			_, err := pipeWriter.Write(rxBuffer.Slice(0, i/phasorSize))
			if err != nil {
				log.Printf("lime: failed to write rx buffer: %s", err)
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

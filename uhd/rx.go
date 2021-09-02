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

package uhd

// #cgo pkg-config: uhd
//
// #include <uhd.h>
import "C"

import (
	"log"
	"unsafe"

	"hz.tools/sdr"
)

// StartRx implements the sdr.Sdr interface.
func (s *Sdr) StartRx() (sdr.ReadCloser, error) {
	var (
		rxStreamerArgs    C.uhd_stream_args_t
		rxStreamer        C.uhd_rx_streamer_handle
		rxMetadata        C.uhd_rx_metadata_handle
		rxStreamerChanLen = C.size_t(1)
		rxStreamerChans   = (*C.size_t)(C.malloc(C.size_t(unsafe.Sizeof(C.size_t(0) * rxStreamerChanLen))))

		streamCmd C.uhd_stream_cmd_t
	)
	*rxStreamerChans = C.size_t(s.rxChannel)
	rxStreamerArgsStr := C.CString("")
	rxStreamFormat := C.CString("sc16")

	unallocRxC := func() {
		C.free(unsafe.Pointer(rxStreamerChans))
		C.free(unsafe.Pointer(rxStreamerArgsStr))
		C.free(unsafe.Pointer(rxStreamFormat))
	}

	if err := rvToError(C.uhd_rx_streamer_make(&rxStreamer)); err != nil {
		unallocRxC()
		return nil, err
	}

	if err := rvToError(C.uhd_rx_metadata_make(&rxMetadata)); err != nil {
		unallocRxC()
		C.uhd_rx_streamer_free(&rxStreamer)
		return nil, err
	}

	unallocRxUhd := func() {
		C.uhd_rx_streamer_free(&rxStreamer)
		C.uhd_rx_metadata_free(&rxMetadata)
	}

	rxStreamerArgs.otw_format = rxStreamFormat
	rxStreamerArgs.cpu_format = rxStreamFormat
	rxStreamerArgs.args = rxStreamerArgsStr
	rxStreamerArgs.channel_list = rxStreamerChans
	rxStreamerArgs.n_channels = C.int(rxStreamerChanLen)

	if err := rvToError(C.uhd_usrp_get_rx_stream(
		*s.handle,
		&rxStreamerArgs,
		rxStreamer,
	)); err != nil {
		unallocRxC()
		unallocRxUhd()
		return nil, err
	}

	sr, err := s.GetSampleRate()
	if err != nil {
		unallocRxC()
		unallocRxUhd()
		return nil, err
	}
	pipeReader, pipeWriter := sdr.Pipe(sr, sdr.SampleFormatI16)

	var (
		iqLength = 1024 * 32 * 16
		iqSize   = iqLength * sdr.SampleFormatI16.Size()
		iqs      = make([]sdr.SamplesI16, 16)
		ciqSize  = C.size_t(iqSize)
		ciqLen   = C.size_t(iqLength)
		ciq      = C.malloc(C.size_t(ciqSize))
	)

	for i := range iqs {
		iqs[i] = make(sdr.SamplesI16, iqLength)
	}

	go func() {
		defer unallocRxC()
		defer unallocRxUhd()
		defer pipeWriter.Close()
		defer C.free(unsafe.Pointer(ciq))

		var (
			n       C.size_t
			i       int
			errCode C.uhd_rx_metadata_error_code_t
		)

		streamCmd.stream_mode = C.UHD_STREAM_MODE_START_CONTINUOUS
		streamCmd.stream_now = true
		if err := rvToError(C.uhd_rx_streamer_issue_stream_cmd(rxStreamer, &streamCmd)); err != nil {
			pipeWriter.CloseWithError(err)
		}

		for {
			i += 1
			if rvToError(C.uhd_rx_streamer_recv(
				rxStreamer, &ciq, ciqLen, &rxMetadata,
				3.0, false, &n,
			)); err != nil {
				pipeWriter.CloseWithError(err)
				return
			}

			if rvToError(C.uhd_rx_metadata_error_code(rxMetadata, &errCode)); err != nil {
				pipeWriter.CloseWithError(err)
				return
			}

			if errCode != C.UHD_RX_METADATA_ERROR_CODE_NONE {
				log.Printf("RX Error: %#v", errCode)
				pipeWriter.Close()
				return
			}

			iq := iqs[i%len(iqs)]
			ciqGB := C.GoBytes(unsafe.Pointer(ciq), C.int(ciqSize))
			copy(sdr.MustUnsafeSamplesAsBytes(iq), ciqGB)

			go func(iq sdr.SamplesI16, n int) {
				iq = iq[:n]

				_, err := pipeWriter.Write(iq)
				if err != nil {
					pipeWriter.CloseWithError(err)
					return
				}
			}(iq, int(n))
		}
	}()

	return pipeReader, nil
}

// vim: foldmethod=marker

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
	"unsafe"

	"hz.tools/sdr"
)

type readCloser struct {
	writer sdr.PipeWriter
	reader sdr.PipeReader
	sdr    *Sdr
	buf    sdr.SamplesI16
}

func (rc *readCloser) Read(iq sdr.Samples) (int, error) {
	return rc.reader.Read(iq)
}

func (rc *readCloser) Close() error {
	rc.writer.Close()
	return nil
}

func (rc *readCloser) SampleRate() uint {
	return rc.reader.SampleRate()
}

func (rc *readCloser) SampleFormat() sdr.SampleFormat {
	return rc.reader.SampleFormat()
}

// StartRx implements the sdr.Sdr interface.
func (s *Sdr) StartRx() (sdr.ReadCloser, error) {
	var (
		args        C.uhd_stream_args_t
		channelList []C.size_t = []C.size_t{C.size_t(s.rxChannel)}

		rxStream C.uhd_rx_streamer_handle
		rxMeta   C.uhd_rx_metadata_handle

		streamCmd C.uhd_stream_cmd_t
	)

	sr, err := s.GetSampleRate()
	if err != nil {
		return nil, err
	}

	format := C.CString("sc16")
	rxArgs := C.CString("")
	defer C.free(unsafe.Pointer(format))
	defer C.free(unsafe.Pointer(rxArgs))

	args.cpu_format = format
	args.otw_format = format
	args.args = rxArgs
	args.channel_list = &channelList[0]
	args.n_channels = 1

	streamCmd.stream_mode = C.UHD_STREAM_MODE_START_CONTINUOUS
	streamCmd.stream_now = true

	pipeReader, pipeWriter := sdr.Pipe(sr, sdr.SampleFormatI16)

	if err := rvToError(C.uhd_rx_streamer_make(&rxStream)); err != nil {
		return nil, err
	}

	if err := rvToError(C.uhd_rx_metadata_make(&rxMeta)); err != nil {
		C.uhd_rx_streamer_free(&rxStream)
		return nil, err
	}

	if err := rvToError(C.uhd_usrp_get_rx_stream(
		*s.handle,
		&args,
		rxStream,
	)); err != nil {
		C.uhd_rx_streamer_free(&rxStream)
		C.uhd_rx_metadata_free(&rxMeta)
		return nil, err
	}

	// var (
	// 	cBuf    *C.float
	// 	cBufLen int = 1024 * 32 * 32
	// )

	// cBuf = (*C.float)(C.malloc(C.size_t(cBufLen) * 2 * C.size_t(unsafe.Sizeof(C.float(0)))))

	// buf := C.GoBytes(unsafe.Pointer(cBuf), C.int(cBufLen))
	// samples := make(sdr.SamplesU8, len(buf)/2)

	// copy(sdr.MustUnsafeSamplesAsBytes(samples), buf)

	go func() {
		defer pipeWriter.Close()
		defer C.uhd_rx_streamer_free(&rxStream)
		defer C.uhd_rx_metadata_free(&rxMeta)

	}()

	return pipeReader, nil
}

// vim: foldmethod=marker

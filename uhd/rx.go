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
	"context"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"hz.tools/sdr"
)

type uhdRxMetadataError int

// Error implements the error type.
func (u uhdRxMetadataError) Error() string {
	switch u {
	case ErrRxMetadataTimeout:
		return "UHD RX Metadata: Timeout"
	case ErrRxMetadataLateCommand:
		return "UHD RX Metadata: Late Command"
	case ErrRxMetadataBrokenChain:
		return "UHD RX Metadata: Broken Chain"
	case ErrRxMetadataOverflow:
		return "UHD RX Metadata: Overflow"
	case ErrRxMetadataAlignment:
		return "UHD RX Metadata: Alignment Error"
	case ErrRxMetadataBadPacket:
		return "UHD RX Metadata: Bad Packet"
	default:
		return "UNKNOWN"
	}
}

var (
	// ErrRxMetadataTimeout will be returned if there's an RX Metadata
	// error condition indicating a timeout.
	ErrRxMetadataTimeout uhdRxMetadataError = 0x01

	// ErrRxMetadataLateCommand will be returned if there's an RX Metadata
	// error condition indicating late command.
	ErrRxMetadataLateCommand uhdRxMetadataError = 0x02

	// ErrRxMetadataBrokenChain will be returned if there's an RX Metadata
	// error condition indicating a broken chain.
	ErrRxMetadataBrokenChain uhdRxMetadataError = 0x04

	// ErrRxMetadataOverflow will be returned if there's an RX Metadata
	// error condition indicating an overflow.
	ErrRxMetadataOverflow uhdRxMetadataError = 0x08

	// ErrRxMetadataAlignment will be returned if there's an RX Metadata
	// error condition indicating a problem with alignment.
	ErrRxMetadataAlignment uhdRxMetadataError = 0x0C

	// ErrRxMetadataBadPacket will be returned if there's an RX Metadata
	// error condition indicating a bad packet.
	ErrRxMetadataBadPacket uhdRxMetadataError = 0x0F
)

// readCloser contains all the allocated structs to be used by the reader
// goroutine and close function.
//
// Most of this stuff isn't stuff that really belongs in here, but the
// allocation lifecycle needs to be tied to this struct.
type readCloser struct {
	closed bool
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	reader       sdr.PipeReader
	writer       sdr.PipeWriter
	sampleFormat sdr.SampleFormat

	rxStreamer C.uhd_rx_streamer_handle
	rxMetadata C.uhd_rx_metadata_handle

	timing struct {
		Set    bool
		Offset time.Duration
	}
}

// Read implements the sdr.Reader interface
func (rc *readCloser) Read(iq sdr.Samples) (int, error) {
	return rc.reader.Read(iq)
}

// SampleRate implements the sdr.Reader interface
func (rc *readCloser) SampleRate() uint {
	return rc.reader.SampleRate()
}

// SampleFormat implements the sdr.Reader interface
func (rc *readCloser) SampleFormat() sdr.SampleFormat {
	return rc.sampleFormat
}

// Close implements the sdr.ReadCloser interface
func (rc *readCloser) Close() error {
	if rc.closed {
		// Avoid double-free'ing or issuing a stream command if we've been
		// called before. This is really a bug, but we wanna be fairly
		// defensive here.
		return nil
	}

	var streamCmd C.uhd_stream_cmd_t

	rc.cancel()
	rc.reader.Close()

	streamCmd.stream_mode = C.UHD_STREAM_MODE_STOP_CONTINUOUS
	streamCmd.stream_now = false
	C.uhd_rx_streamer_issue_stream_cmd(rc.rxStreamer, &streamCmd)

	// Wait until the STOP command has gone through and we're sure the
	// goroutine is stopped. This means that we can free the resources below,
	// otherwise we risk a SEGV.
	rc.wg.Wait()

	C.uhd_rx_streamer_free(&rc.rxStreamer)
	C.uhd_rx_metadata_free(&rc.rxMetadata)

	// TODO(paultag): Literally any error checking at all :)

	rc.closed = true
	return nil
}

// run is a goroutine to handle copying IQ data from the UHD device
// to the Pipe contained inside the readCloser.
func (rc *readCloser) run() {
	defer rc.writer.Close()
	defer rc.cancel()
	defer rc.wg.Done()

	var ciqLen C.size_t

	if err := rvToError(C.uhd_rx_streamer_max_num_samps(rc.rxStreamer, &ciqLen)); err != nil {
		rc.writer.CloseWithError(err)
		return
	}

	var (
		n         C.size_t
		i         int
		errCode   C.uhd_rx_metadata_error_code_t
		streamCmd C.uhd_stream_cmd_t
		err       error

		iqLength = int(ciqLen)
		iqSize   = iqLength * rc.sampleFormat.Size()
		ciqSize  = C.size_t(iqSize)
		ciq      = C.malloc(C.size_t(ciqSize))
	)

	iq, err := sdr.MakeSamples(rc.sampleFormat, iqLength)
	if err != nil {
		rc.writer.CloseWithError(err)
		return
	}

	var hasTimeSpec = C.bool(rc.timing.Set)
	secs, frac := splitDuration(rc.timing.Offset)

	streamCmd.stream_mode = C.UHD_STREAM_MODE_START_CONTINUOUS
	streamCmd.stream_now = !hasTimeSpec
	streamCmd.time_spec_full_secs = secs
	streamCmd.time_spec_frac_secs = frac

	if err := rvToError(C.uhd_rx_streamer_issue_stream_cmd(rc.rxStreamer, &streamCmd)); err != nil {
		rc.writer.CloseWithError(err)
	}

	for {
		i++
		if err := rc.ctx.Err(); err != nil {
			return
		}

		if rvToError(C.uhd_rx_streamer_recv(
			rc.rxStreamer, &ciq, ciqLen, &rc.rxMetadata,
			3.0, false, &n,
		)); err != nil {
			rc.writer.CloseWithError(err)
			return
		}

		if rvToError(C.uhd_rx_metadata_error_code(rc.rxMetadata, &errCode)); err != nil {
			rc.writer.CloseWithError(err)
			return
		}

		if errCode != C.UHD_RX_METADATA_ERROR_CODE_NONE {
			rc.writer.CloseWithError(uhdRxMetadataError(errCode))
			return
		}

		ciqGB := C.GoBytes(unsafe.Pointer(ciq), C.int(ciqSize))
		copy(sdr.MustUnsafeSamplesAsBytes(iq), ciqGB)

		iq := iq.Slice(0, int(n))
		_, err := rc.writer.Write(iq)
		if err != nil {
			rc.writer.CloseWithError(err)
			return
		}
	}
}

type startRxOpts struct {
	Timing struct {
		Set    bool
		Offset time.Duration
	}
}

// StartRx implements the sdr.Sdr interface.
func (s *Sdr) StartRx() (sdr.ReadCloser, error) {
	opts := startRxOpts{}
	return s.startRx(opts)
}

// StartRxAt will StartRx at the specific time offset.
func (s *Sdr) StartRxAt(d time.Duration) (sdr.ReadCloser, error) {
	opts := startRxOpts{}
	opts.Timing.Set = true
	opts.Timing.Offset = d
	return s.startRx(opts)
}

func (s *Sdr) startRx(opts startRxOpts) (sdr.ReadCloser, error) {

	// Before we get down the road of allocating anything, let's check
	// to ensure that we have a supported SampleFormat.
	var format string
	switch s.sampleFormat {
	case sdr.SampleFormatI8:
		format = "sc8"
	case sdr.SampleFormatI16:
		format = "sc16"
	case sdr.SampleFormatC64:
		format = "fc32"
	default:
		return nil, fmt.Errorf("uhd: StartRx: unsupported SampleFormat provided")
	}

	var (
		rxStreamerArgs    C.uhd_stream_args_t
		rxStreamer        C.uhd_rx_streamer_handle
		rxMetadata        C.uhd_rx_metadata_handle
		rxStreamerChanLen = C.size_t(1)
		rxStreamerChans   = (*C.size_t)(C.malloc(C.size_t(unsafe.Sizeof(C.size_t(0) * rxStreamerChanLen))))
	)

	ctx, cancel := context.WithCancel(context.Background())

	*rxStreamerChans = C.size_t(s.rxChannel)
	rxStreamerArgsStr := C.CString("")
	rxStreamFormat := C.CString(format)

	// TODO(paultag): Is it safe to free these even though they were passed
	// into a constructor for the rx streamer?
	//
	// It's my assumption that they're copied in if they're used outside the
	// constructor; but that needs to be validated. This doesn't obviously crash,
	// and this makes the readCloser significantly easier to maintain, and
	// the error cases in the constructor here a lot easier too.
	defer C.free(unsafe.Pointer(rxStreamerChans))
	defer C.free(unsafe.Pointer(rxStreamerArgsStr))
	defer C.free(unsafe.Pointer(rxStreamFormat))

	if err := rvToError(C.uhd_rx_streamer_make(&rxStreamer)); err != nil {
		return nil, err
	}

	if err := rvToError(C.uhd_rx_metadata_make(&rxMetadata)); err != nil {
		C.uhd_rx_streamer_free(&rxStreamer)
		return nil, err
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
		C.uhd_rx_streamer_free(&rxStreamer)
		C.uhd_rx_metadata_free(&rxMetadata)
		return nil, err
	}

	sr, err := s.GetSampleRate()
	if err != nil {
		C.uhd_rx_streamer_free(&rxStreamer)
		C.uhd_rx_metadata_free(&rxMetadata)
		return nil, err
	}

	// TODO(paultag): dynamic SampleFormat
	pipeReader, pipeWriter := sdr.PipeWithContext(ctx, sr, s.sampleFormat)

	rc := &readCloser{
		wg:     sync.WaitGroup{},
		ctx:    ctx,
		cancel: cancel,

		sampleFormat: s.sampleFormat,
		reader:       pipeReader,
		writer:       pipeWriter,

		rxStreamer: rxStreamer,
		rxMetadata: rxMetadata,
	}
	rc.timing.Set = opts.Timing.Set
	rc.timing.Offset = opts.Timing.Offset
	rc.wg.Add(1)
	go rc.run()
	return rc, nil
}

// vim: foldmethod=marker

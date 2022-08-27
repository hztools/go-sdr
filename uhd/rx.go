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
	"hz.tools/sdr/internal/yikes"
	"hz.tools/sdr/stream"
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

// readStreamer contains all the allocated structs to be used by the reader
// goroutine and close function.
//
// Most of this stuff isn't stuff that really belongs in here, but the
// allocation lifecycle needs to be tied to this struct.
type readStreamer struct {
	closed bool
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	writers      pipeWriters
	sampleFormat sdr.SampleFormat

	rxStreamer C.uhd_rx_streamer_handle
	rxMetadata C.uhd_rx_metadata_handle

	timing struct {
		Set    bool
		Offset time.Duration
	}
}

type pipeWriters []*stream.BufPipe2

func (pr pipeWriters) CloseWithError(e error) error {
	var ret error
	for _, el := range pr {
		if err := el.CloseWithError(e); err != nil {
			ret = err
		}
	}
	return ret
}

func (pr pipeWriters) Close() error {
	var ret error
	for _, el := range pr {
		if err := el.Close(); err != nil {
			ret = err
		}
	}
	return ret
}

// Close implements the sdr.ReadCloser interface
func (rc *readStreamer) Close() error {
	if rc.closed {
		// Avoid double-free'ing or issuing a stream command if we've been
		// called before. This is really a bug, but we wanna be fairly
		// defensive here.
		return nil
	}

	var streamCmd C.uhd_stream_cmd_t

	rc.cancel()
	rc.writers.Close()

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
// to the Pipe contained inside the readStreamer.
func (rc *readStreamer) run() {
	defer rc.writers.Close()
	defer rc.cancel()
	defer rc.wg.Done()

	var ciqLen C.size_t
	if err := rvToError(C.uhd_rx_streamer_max_num_samps(rc.rxStreamer, &ciqLen)); err != nil {
		rc.writers.CloseWithError(err)
		return
	}

	var channels = len(rc.writers)
	if channels > 32 {
		panic("UHD: too many rx channels set")
	}

	var (
		n         C.size_t
		i         int
		errCode   C.uhd_rx_metadata_error_code_t
		streamCmd C.uhd_stream_cmd_t

		iqLength = int(ciqLen)
		iqSize   = iqLength * rc.sampleFormat.Size()
		ciqSize  = C.size_t(iqSize)

		cIQBuffers = make([]unsafe.Pointer, channels)
	)
	for i := 0; i < len(rc.writers); i++ {
		cIQBuffers[i] = C.malloc(C.size_t(ciqSize))
	}

	var hasTimeSpec = C.bool(rc.timing.Set)
	secs, frac := splitDuration(rc.timing.Offset)

	streamCmd.stream_mode = C.UHD_STREAM_MODE_START_CONTINUOUS
	streamCmd.stream_now = !hasTimeSpec
	streamCmd.time_spec_full_secs = secs
	streamCmd.time_spec_frac_secs = frac

	if err := rvToError(C.uhd_rx_streamer_issue_stream_cmd(rc.rxStreamer, &streamCmd)); err != nil {
		rc.writers.CloseWithError(err)
	}

	for {
		i++
		if err := rc.ctx.Err(); err != nil {
			return
		}

		if err := rvToError(C.uhd_rx_streamer_recv(
			rc.rxStreamer, &cIQBuffers[0], ciqLen, &rc.rxMetadata,
			3.0, false, &n,
		)); err != nil {
			rc.writers.CloseWithError(err)
			return
		}

		if err := rvToError(C.uhd_rx_metadata_error_code(rc.rxMetadata, &errCode)); err != nil {
			rc.writers.CloseWithError(err)
			return
		}

		if errCode != C.UHD_RX_METADATA_ERROR_CODE_NONE {
			rc.writers.CloseWithError(uhdRxMetadataError(errCode))
			return
		}

		for i := 0; i < channels; i++ {
			ciq := cIQBuffers[i]
			writer := rc.writers[i]
			iq, err := yikes.Samples(uintptr(ciq), iqLength, rc.sampleFormat)
			if err != nil {
				rc.writers.CloseWithError(uhdRxMetadataError(errCode))
				return
			}
			iq = iq.Slice(0, int(n))
			_, err = writer.Write(iq)
			if err != nil {
				rc.writers.CloseWithError(err)
				return
			}
		}
	}
}

type startRxOpts struct {
	BufferLength int
	RxChannels   []int
	Timing       struct {
		Set    bool
		Offset time.Duration
	}
}

// StartRx implements the sdr.Sdr interface.
func (s *Sdr) StartRx() (sdr.ReadCloser, error) {
	if len(s.rxChannels) != 1 {
		return nil, fmt.Errorf("uhd: rx: only one channel can be provided")
	}

	opts := startRxOpts{
		BufferLength: s.bufferLength,
		RxChannels:   s.rxChannels,
	}
	rcs, err := s.startRx(opts)
	if err != nil {
		return nil, err
	}
	return rcs[0], nil
}

// StartRxAt will StartRx at the specific time offset.
func (s *Sdr) StartRxAt(d time.Duration) (sdr.ReadCloser, error) {
	if len(s.rxChannels) != 1 {
		return nil, fmt.Errorf("uhd: rx: only one channel can be provided")
	}

	opts := startRxOpts{
		BufferLength: s.bufferLength,
		RxChannels:   s.rxChannels,
	}
	opts.Timing.Set = true
	opts.Timing.Offset = d
	rcs, err := s.startRx(opts)
	if err != nil {
		return nil, err
	}
	return rcs[0], nil
}

// StartCoherentRx will start a coherent RX operation. As a byproduct, this
// will reset the clock.
func (s *Sdr) StartCoherentRx() (sdr.ReadClosers, error) {
	if err := s.SetTimeNow(time.Duration(0)); err != nil {
		return nil, err
	}
	return s.StartCoherentRxAt(time.Second)
}

// StartCoherentRxAt will start a coherent RX operation, sync'd at the
// provided offset.
func (s *Sdr) StartCoherentRxAt(d time.Duration) (sdr.ReadClosers, error) {
	opts := startRxOpts{
		BufferLength: s.bufferLength,
		RxChannels:   s.rxChannels,
	}
	opts.Timing.Set = true
	opts.Timing.Offset = d
	return s.startRx(opts)
}

func (s *Sdr) startRx(opts startRxOpts) (sdr.ReadClosers, error) {
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

	channels := len(opts.RxChannels)
	if channels > 32 {
		return nil, fmt.Errorf("uhd: wow, that's a lot of channels; this breaks some internals, please fix")
	}

	var (
		rxStreamerArgs    C.uhd_stream_args_t
		rxStreamer        C.uhd_rx_streamer_handle
		rxMetadata        C.uhd_rx_metadata_handle
		rxStreamerChanLen = C.size_t(channels)
		rxStreamerChans   = (*C.size_t)(C.malloc(C.size_t(unsafe.Sizeof(C.size_t(0)) * uintptr(channels))))
		rxStreamerGoChans = (*[1 << 30]C.size_t)(unsafe.Pointer(rxStreamerChans))[:channels:channels]
	)

	ctx, cancel := context.WithCancel(context.Background())
	for i, c := range opts.RxChannels {
		rxStreamerGoChans[i] = C.size_t(c)
	}

	rxStreamerArgsStr := C.CString("")
	rxStreamFormat := C.CString(format)

	// TODO(paultag): Is it safe to free these even though they were passed
	// into a constructor for the rx streamer?
	//
	// It's my assumption that they're copied in if they're used outside the
	// constructor; but that needs to be validated. This doesn't obviously crash,
	// and this makes the readStreamer significantly easier to maintain, and
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

	bufferLength := opts.BufferLength

	pipes := make([]*stream.BufPipe2, len(opts.RxChannels))
	readers := make(sdr.ReadClosers, len(opts.RxChannels))

	for i := range opts.RxChannels {
		// TODO(paultag): 10 isn't right here.
		pipe, err := stream.NewBufPipe2(bufferLength, sr, s.sampleFormat)
		if err != nil {
			return nil, err
		}
		pipes[i] = pipe
		readers[i] = pipe
	}

	rc := &readStreamer{
		wg:     sync.WaitGroup{},
		ctx:    ctx,
		cancel: cancel,

		sampleFormat: s.sampleFormat,
		writers:      pipes,

		rxStreamer: rxStreamer,
		rxMetadata: rxMetadata,
	}
	rc.timing.Set = opts.Timing.Set
	rc.timing.Offset = opts.Timing.Offset
	rc.wg.Add(1)
	go rc.run()
	return readers, nil
}

// vim: foldmethod=marker

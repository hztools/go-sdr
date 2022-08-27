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

// writeCloser contains all the allocated structs to be used by the writeer
// goroutine and close function.
//
// Most of this stuff isn't stuff that really belongs in here, but the
// allocation lifecycle needs to be tied to this struct.
type writeCloser struct {
	closed bool
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	pipe         *stream.BufPipe2
	sampleFormat sdr.SampleFormat

	txStreamer C.uhd_tx_streamer_handle
	txMetadata C.uhd_tx_metadata_handle
}

// Write implements the sdr.Writer interface
func (wc *writeCloser) Write(iq sdr.Samples) (int, error) {
	return wc.pipe.Write(iq)
}

// SampleRate implements the sdr.Writer interface
func (wc *writeCloser) SampleRate() uint {
	return wc.pipe.SampleRate()
}

// SampleFormat implements the sdr.Writer interface
func (wc *writeCloser) SampleFormat() sdr.SampleFormat {
	return wc.sampleFormat
}

// Close implements the sdr.WriteCloser interface
func (wc *writeCloser) Close() error {
	if wc.closed {
		// Avoid double-free'ing or issuing a stream command if we've been
		// called before. This is really a bug, but we wanna be fairly
		// defensive here.
		return nil
	}

	wc.pipe.Close()
	wc.cancel()

	// Wait until pipe is read fully, and we're sure the goroutine is stopped.
	// This means that we can free the resouwces below, otherwise we risk a
	// SEGV.
	wc.wg.Wait()

	C.uhd_tx_streamer_free(&wc.txStreamer)
	C.uhd_tx_metadata_free(&wc.txMetadata)

	// TODO(paultag): Literally any error checking at all :)

	wc.closed = true
	return nil
}

// run is a goroutine to handle copying IQ data from the UHD device
// to the Pipe contained inside the writeCloser.
func (wc *writeCloser) run() {
	defer wc.pipe.Close()
	defer wc.cancel()
	defer wc.wg.Done()

	var ciqLen C.size_t

	if err := rvToError(C.uhd_tx_streamer_max_num_samps(wc.txStreamer, &ciqLen)); err != nil {
		wc.pipe.CloseWithError(err)
		return
	}

	var (
		cn  C.size_t
		i   int
		err error

		iqLength = int(ciqLen)
		iqSize   = iqLength * wc.sampleFormat.Size()
		ciqSize  = C.size_t(iqSize)
		ciq      = C.malloc(C.size_t(ciqSize))
	)

	iq, err := sdr.MakeSamples(wc.sampleFormat, iqLength)
	if err != nil {
		wc.pipe.CloseWithError(err)
		return
	}

	// Blank out the C memory
	copy(yikes.GoBytes(uintptr(unsafe.Pointer(ciq)), iqSize),
		sdr.MustUnsafeSamplesAsBytes(iq))

	// before we do anything, let's send a buffer to let
	// the hardware warm up and get something to chew on
	// while we get going here
	for i := 0; i < 10; i++ {
		if err := rvToError(C.uhd_tx_streamer_send(
			wc.txStreamer, &ciq, ciqLen, &wc.txMetadata,
			0.1, &cn,
		)); err != nil {
			return
		}
	}

	for {
		i++
		n, rferr := sdr.ReadFull(wc.pipe, iq)

		if n != iq.Length() {
			if rferr == nil {
				// this is bad, something is broken
				wc.pipe.CloseWithError(fmt.Errorf("uhd: ReadFull was short"))
				return
			}
		}

		copy(yikes.GoBytes(uintptr(unsafe.Pointer(ciq)), iqSize),
			sdr.MustUnsafeSamplesAsBytes(iq))

		if err := rvToError(C.uhd_tx_streamer_send(
			wc.txStreamer, &ciq, C.size_t(n), &wc.txMetadata,
			0.1, &cn,
		)); err != nil {
			// wc.pipe.CloseWithError(err)
			return
		}

		if rferr != nil {
			// if our ReadFull had an error, we can bail now that we sent
			// the last windowsworth.
			return
		}
	}
}

// StartTxAt will start TX at the provided Duration offset.
func (s *Sdr) StartTxAt(d time.Duration) (sdr.WriteCloser, error) {
	opts := startTxOpts{BufferLength: s.bufferLength}
	opts.Timing.Set = true
	opts.Timing.Offset = d
	return s.startTx(opts)
}

// StartTx implements the sdr.Sdr interface.
func (s *Sdr) StartTx() (sdr.WriteCloser, error) {
	opts := startTxOpts{BufferLength: s.bufferLength}
	return s.startTx(opts)
}

type startTxOpts struct {
	BufferLength int
	Timing       struct {
		Set    bool
		Offset time.Duration
	}
}

func (s *Sdr) startTx(opts startTxOpts) (sdr.WriteCloser, error) {
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
		txStreamerArgs    C.uhd_stream_args_t
		txStreamer        C.uhd_tx_streamer_handle
		txMetadata        C.uhd_tx_metadata_handle
		txStreamerChanLen = C.size_t(1)
		txStreamerChans   = (*C.size_t)(C.malloc(C.size_t(unsafe.Sizeof(C.size_t(0) * txStreamerChanLen))))
	)

	ctx, cancel := context.WithCancel(context.Background())

	*txStreamerChans = C.size_t(s.txChannel)
	txStreamerArgsStr := C.CString("")
	txStreamFormat := C.CString(format)

	// TODO(paultag): Is it safe to free these even though they were passed
	// into a constructor for the tx streamer?
	//
	// It's my assumption that they're copied in if they're used outside the
	// constructor; but that needs to be validated. This doesn't obviously crash,
	// and this makes the readCloser significantly easier to maintain, and
	// the error cases in the constructor here a lot easier too.
	defer C.free(unsafe.Pointer(txStreamerChans))
	defer C.free(unsafe.Pointer(txStreamerArgsStr))
	defer C.free(unsafe.Pointer(txStreamFormat))

	if err := rvToError(C.uhd_tx_streamer_make(&txStreamer)); err != nil {
		return nil, err
	}

	var hasTimeSpec = C.bool(opts.Timing.Set)
	secs, frac := splitDuration(opts.Timing.Offset)

	if err := rvToError(C.uhd_tx_metadata_make(&txMetadata, hasTimeSpec, secs, frac, true, false)); err != nil {
		C.uhd_tx_streamer_free(&txStreamer)
		return nil, err
	}

	txStreamerArgs.otw_format = txStreamFormat
	txStreamerArgs.cpu_format = txStreamFormat
	txStreamerArgs.args = txStreamerArgsStr
	txStreamerArgs.channel_list = txStreamerChans
	txStreamerArgs.n_channels = C.int(txStreamerChanLen)

	if err := rvToError(C.uhd_usrp_get_tx_stream(
		*s.handle,
		&txStreamerArgs,
		txStreamer,
	)); err != nil {
		C.uhd_tx_streamer_free(&txStreamer)
		C.uhd_tx_metadata_free(&txMetadata)
		return nil, err
	}

	sr, err := s.GetSampleRate()
	if err != nil {
		C.uhd_tx_streamer_free(&txStreamer)
		C.uhd_tx_metadata_free(&txMetadata)
		return nil, err
	}

	bufferLength := opts.BufferLength

	bp, err := stream.NewBufPipe2(bufferLength, sr, s.sampleFormat)
	if err != nil {
		C.uhd_tx_streamer_free(&txStreamer)
		C.uhd_tx_metadata_free(&txMetadata)
		return nil, err
	}

	wc := &writeCloser{
		wg:     sync.WaitGroup{},
		ctx:    ctx,
		cancel: cancel,

		sampleFormat: s.sampleFormat,
		pipe:         bp,

		txStreamer: txStreamer,
		txMetadata: txMetadata,
	}
	wc.wg.Add(1)
	go wc.run()
	return wc, nil
}

// vim: foldmethod=marker

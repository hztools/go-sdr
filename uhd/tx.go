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
	"unsafe"

	"hz.tools/sdr"
	"hz.tools/sdr/internal/yikes"
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

	reader       sdr.PipeReader
	writer       sdr.PipeWriter
	sampleFormat sdr.SampleFormat

	txStreamer C.uhd_tx_streamer_handle
	txMetadata C.uhd_tx_metadata_handle
}

// Write implements the sdr.Writer interface
func (rc *writeCloser) Write(iq sdr.Samples) (int, error) {
	return rc.writer.Write(iq)
}

// SampleRate implements the sdr.Writer interface
func (rc *writeCloser) SampleRate() uint {
	return rc.writer.SampleRate()
}

// SampleFormat implements the sdr.Writer interface
func (rc *writeCloser) SampleFormat() sdr.SampleFormat {
	return rc.sampleFormat
}

// Close implements the sdr.WriteCloser interface
func (rc *writeCloser) Close() error {
	if rc.closed {
		// Avoid double-free'ing or issuing a stream command if we've been
		// called before. This is really a bug, but we wanna be fairly
		// defensive here.
		return nil
	}

	rc.cancel()
	rc.writer.Close()

	// Wait until the STOP command has gone through and we're sure the
	// goroutine is stopped. This means that we can free the resources below,
	// otherwise we risk a SEGV.
	rc.wg.Wait()

	C.uhd_tx_streamer_free(&rc.txStreamer)
	C.uhd_tx_metadata_free(&rc.txMetadata)

	// TODO(paultag): Literally any error checking at all :)

	rc.closed = true
	return nil
}

// run is a goroutine to handle copying IQ data from the UHD device
// to the Pipe contained inside the writeCloser.
func (rc *writeCloser) run() {
	defer rc.writer.Close()
	defer rc.cancel()
	defer rc.wg.Done()

	var ciqLen C.size_t

	if err := rvToError(C.uhd_tx_streamer_max_num_samps(rc.txStreamer, &ciqLen)); err != nil {
		rc.writer.CloseWithError(err)
		return
	}

	var (
		cn  C.size_t
		i   int
		err error

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

	for {
		i++
		if err := rc.ctx.Err(); err != nil {
			return
		}

		n, err := sdr.ReadFull(rc.reader, iq)
		if err != nil {
			rc.writer.CloseWithError(err)
			return
		}
		cn = C.size_t(n)

		ciqGB := yikes.GoBytes(uintptr(unsafe.Pointer(ciq)), iqSize)
		copy(ciqGB, sdr.MustUnsafeSamplesAsBytes(iq))

		if rvToError(C.uhd_tx_streamer_send(
			rc.txStreamer, &ciq, ciqLen, &rc.txMetadata,
			0.1, &cn,
		)); err != nil {
			rc.writer.CloseWithError(err)
			return
		}
	}
}

// StartTx implements the sdr.Sdr interface.
func (s *Sdr) StartTx() (sdr.WriteCloser, error) {
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

	if err := rvToError(C.uhd_tx_metadata_make(&txMetadata, false, 0, 0.1, true, false)); err != nil {
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

	// TODO(paultag): dynamic SampleFormat
	pipeReader, pipeWriter := sdr.PipeWithContext(ctx, sr, s.sampleFormat)

	rc := &writeCloser{
		wg:     sync.WaitGroup{},
		ctx:    ctx,
		cancel: cancel,

		sampleFormat: s.sampleFormat,
		reader:       pipeReader,
		writer:       pipeWriter,

		txStreamer: txStreamer,
		txMetadata: txMetadata,
	}
	rc.wg.Add(1)
	go rc.run()
	return rc, nil
}

// vim: foldmethod=marker

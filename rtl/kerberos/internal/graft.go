// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2020
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

package internal

import (
	"context"
	"log"

	"hz.tools/sdr"
	"hz.tools/sdr/fft"
)

type graftReader struct {
	ctx    context.Context
	closer context.CancelFunc

	planner fft.Planner
	readers []sdr.Reader

	fftSize int
	iqBufs  []sdr.SamplesC64

	pipeReader sdr.PipeReader
	pipeWriter sdr.PipeWriter
}

func (gr *graftReader) Read(s sdr.Samples) (int, error) {
	return gr.pipeReader.Read(s)
}

func (gr *graftReader) SampleFormat() sdr.SampleFormat {
	return gr.pipeReader.SampleFormat()
}

func (gr *graftReader) SampleRate() uint {
	return gr.pipeReader.SampleRate()
}

func (gr *graftReader) Close() error {
	gr.closer()
	gr.pipeWriter.Close()
	return nil
}

func (gr *graftReader) do() error {
	outBuf := make(sdr.SamplesC64, gr.fftSize*len(gr.readers))
	freqBuf := make([]complex64, gr.fftSize*len(gr.readers))

	fftPlans := make([]fft.Plan, len(gr.readers))

	for i := range gr.iqBufs {
		start := gr.fftSize * i
		end := (gr.fftSize * (i + 1))
		freqBufI := freqBuf[start:end]
		plan, err := gr.planner(gr.iqBufs[i], freqBufI, fft.Forward)
		if err != nil {
			return err
		}
		fftPlans[i] = plan
	}

	fftOutPlan, err := gr.planner(outBuf, freqBuf, fft.Backward)
	if err != nil {
		return err
	}

	for {
		if err := gr.ctx.Err(); err != nil {
			return err
		}

		if err := ReadBuffers(
			gr.readers,
			gr.iqBufs,
		); err != nil {
			return err
		}

		// for i := 2; i <= 3; i++ {
		for i := range gr.iqBufs {
			start := gr.fftSize * i
			end := (gr.fftSize * (i + 1))
			freqBufI := freqBuf[start:end]

			if err := fftPlans[i].Transform(); err != nil {
				return err
			}

			if err := FFTShiftAndScale(freqBufI, float32(gr.fftSize)); err != nil {
				return err
			}
		}

		if err := fftOutPlan.Transform(); err != nil {
			return err
		}

		_, err := gr.pipeWriter.Write(outBuf)
		if err != nil {
			return err
		}
	}
	return nil
}

// GraftReaders will combine (in frequency space) multiple readers, and return
// an sdr.Reader at a higher sample rate.
func GraftReaders(planner fft.Planner, readers []sdr.Reader) (sdr.ReadCloser, error) {
	var (
		bufl   = 1024 * 64
		lenr   = len(readers)
		iqBufs = make([]sdr.SamplesC64, lenr)

		sampleRate uint = uint(len(readers)) * readers[0].SampleRate()
	)

	for i := range iqBufs {
		iqBufs[i] = make(sdr.SamplesC64, bufl)
	}

	ctx, closer := context.WithCancel(context.Background())

	pipeReader, pipeWriter := sdr.Pipe(sampleRate, sdr.SampleFormatC64)

	gr := &graftReader{
		ctx:        ctx,
		closer:     closer,
		planner:    planner,
		readers:    readers,
		fftSize:    bufl,
		iqBufs:     iqBufs,
		pipeReader: pipeReader,
		pipeWriter: pipeWriter,
	}
	go func() {
		if err := gr.do(); err != nil {
			log.Printf("Error grafting: %s\n", err)
		}
	}()
	return gr, nil
}

// vim: foldmethod=marker

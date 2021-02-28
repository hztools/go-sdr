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

package stream

import (
	"hz.tools/sdr"
)

type readTransformer struct {
	sdr.PipeReader

	input       sdr.Reader
	inputBuffer sdr.Samples

	output       sdr.PipeWriter
	outputBuffer sdr.Samples

	config ReadTransformerConfig
}

// ReadTransformer will wrap an sdr.Pipe and run the provided Processer over
// each chunk as available.
//
// This code will spawn a single Goroutine for each invocation who's lifetime
// is tied to the provided Reader. Be sure to close the reader or this code
// will remain alive forever!
func ReadTransformer(in sdr.Reader, config ReadTransformerConfig) (sdr.Reader, error) {
	pipeReader, pipeWriter := sdr.Pipe(
		config.OutputSampleRate,
		config.OutputSampleFormat,
	)

	inputBuffer, err := sdr.MakeSamples(
		in.SampleFormat(),
		config.InputBufferLength,
	)
	if err != nil {
		return nil, err
	}

	outputBuffer, err := sdr.MakeSamples(
		config.OutputSampleFormat,
		config.OutputBufferLength,
	)
	if err != nil {
		return nil, err
	}

	r := &readTransformer{
		PipeReader: pipeReader,

		input:        in,
		inputBuffer:  inputBuffer,
		output:       pipeWriter,
		outputBuffer: outputBuffer,

		config: config,
	}

	// if err := r.check(); err != nil {
	// 	return nil, err
	// }

	go r.run()

	return r, nil
}

// ReadTransformerConfig defines how the ReadTransformer will process samples,
// and what types of buffers to create.
//
// This is intended to take care of fairly generic cases where a bit of code
// transforms the *read* path of the IQ data.
type ReadTransformerConfig struct {
	// InputBufferLength will dictate the size of the Input buffer.
	InputBufferLength int

	// OutputBufferLength will dictate the size of the Output buffer.
	OutputBufferLength int

	// OutputSampleFormat will define the sample format of the output Reader.
	// This won't always be the same format as the input stream (such as
	// in the case of Conversion)
	OutputSampleFormat sdr.SampleFormat

	// OutputSampleRate will define the sample rate of the output Reader.
	// This won't always be the same rate as the input stream (such as
	// in the case of Decimation).
	OutputSampleRate uint32

	// Proc is the actual code to process each chunk of IQ data.
	//
	// This must return the number of samples written to the output buffer.
	// Any errors will be set on the underlying Reader, so that future calls to
	// the Reader returned by the initial ReadTransformer call will retun that
	// error.
	Proc func(in sdr.Samples, out sdr.Samples) (int, error)
}

func (r *readTransformer) run() error {
	defer r.output.Close()
	for {
		inn, err := r.input.Read(r.inputBuffer)
		if err != nil {
			r.output.CloseWithError(err)
			return err
		}
		outn, err := r.config.Proc(r.inputBuffer.Slice(0, inn), r.outputBuffer)
		if err != nil {
			r.output.CloseWithError(err)
			return err
		}
		_, err = r.output.Write(r.outputBuffer.Slice(0, outn))
		if err != nil {
			r.output.CloseWithError(err)
			return err
		}
	}
}

// vim: foldmethod=marker

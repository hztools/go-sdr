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

package sdr

// Readers represents a collection of Readers.
type Readers []Reader

// SampleRate returns the number of samples per second in the stream.
func (rs Readers) SampleRate() uint {
	if len(rs) == 0 {
		return 0
	}
	ret := rs[0].SampleRate()
	for _, r := range rs {
		if r.SampleRate() != ret {
			return 0
		}
	}
	return ret
}

// SampleFormat returns the IQ Format of the Readers.
func (rs Readers) SampleFormat() SampleFormat {
	if len(rs) == 0 {
		return SampleFormat(0)
	}
	ret := rs[0].SampleFormat()
	for _, r := range rs {
		if r.SampleFormat() != ret {
			return SampleFormat(0)
		}
	}
	return ret
}

// WrapErr will apply the function 'fn' to each Reader in this collection,
// and return a new slice of those new Reader objects. If any error is
// encountered, that error is returned.
func (rs Readers) WrapErr(fn func(Reader) (Reader, error)) (Readers, error) {
	var (
		err error
		ret = make(Readers, len(rs))
	)
	for i := range rs {
		ret[i], err = fn(rs[i])
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

// Wrap will apply the function 'fn' to each Reader in this collection,
// and return a new slice of those new Reader objects.
func (rs Readers) Wrap(fn func(Reader) Reader) Readers {
	ret := make(Readers, len(rs))
	for i := range rs {
		ret[i] = fn(rs[i])
	}
	return ret
}

// ReadClosers is a collection of ReadCloser objects.
type ReadClosers []ReadCloser

// SampleRate returns the number of IQ samples per second.
func (rcs ReadClosers) SampleRate() uint {
	return rcs.Readers().SampleRate()
}

// SampleFormat returns the IQ format of the Readers.
func (rcs ReadClosers) SampleFormat() SampleFormat {
	return rcs.Readers().SampleFormat()
}

// Readers will return the ReadClosers as a Reader slice.
func (rcs ReadClosers) Readers() Readers {
	ret := make(Readers, len(rcs))
	for i := range rcs {
		ret[i] = rcs[i]
	}
	return ret
}

// Close will close all the ReadClosers.
func (rcs ReadClosers) Close() error {
	for _, rc := range rcs {
		if err := rc.Close(); err != nil {
			return err
		}
	}
	return nil
}

// vim: foldmethod=marker

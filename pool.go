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

import (
	"sync"
)

// SamplesPool creates a dynamically sized buffer pool of a set size and
// sample format. This allows code to reuse buffers, and avoid allocations
// if a buffer is ready for use.
//
// Under the hood this is a sync.Pool, but with some type-safe (well, as
// type safe as you can get by returning another interface type...) hooks
// to make this a bit more ergonomic to use.
type SamplesPool struct {
	pool *sync.Pool
}

// Put will return a buffer to the pool. In the future this is likely to panic
// if the buffer returned is not the same size as the buffer that was taken out,
// but since the size won't /shrink/, it's actually not that bad to deal with
// here.
func (sp SamplesPool) Put(s Samples) {
	sp.pool.Put(s)
}

// Get will either return an unused buffer, or allocate a new one for your use.
//
// The smallest size of a buffer returned will be the length passed to the
// NewSamplesPool constructor, of the provided SampleFormat.
func (sp SamplesPool) Get() Samples {
	return sp.pool.Get().(Samples)
}

// NewSamplesPool will create a new SamplesPool that creates buffers of the
// provided sample format and length.
//
// Under the hood this is a wrapped sync.Pool.
func NewSamplesPool(format SampleFormat, length int) (*SamplesPool, error) {
	switch format {
	case SampleFormatU8, SampleFormatI8, SampleFormatI16, SampleFormatC64:
		break
	default:
		return nil, ErrSampleFormatUnknown
	}

	return &SamplesPool{
		pool: &sync.Pool{
			New: func() interface{} {
				// Ouch. Hopefully the type check above helps avoid
				// other errors at runtime...
				buf, _ := MakeSamples(format, length)
				return buf
			},
		},
	}, nil
}

// vim: foldmethod=marker

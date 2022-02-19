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

//go:build !sdr.nosimd
// +build !sdr.nosimd

package simd

import (
	"log"
)

// mmxCPUSupport will run a CPUID check to see if the CPU we're
// running on supports MMX.
func mmxCPUSupport() bool
func avxCPUSupport() bool

func init() {
	mmx := mmxCPUSupport()
	avx := avxCPUSupport()

	if !mmx || !avx {
		log.Printf("[hz.tools/sdr] SIMD support was compiled in, but the CPU\n")
		log.Printf("[hz.tools/sdr] claims to not support MMX, AVX and AVX2. As a result,\n")
		log.Printf("[hz.tools/sdr] I'm going to cowardly refuse to run!\n")
		log.Printf("[hz.tools/sdr]\n")
		log.Printf("[hz.tools/sdr] I suggest attempting to rebuild this binary\n")
		log.Printf("[hz.tools/sdr] using `-tags=sdr.nosimd` to disable SIMD\n")
		log.Printf("[hz.tools/sdr] support in this library.\n")
		log.Printf("[hz.tools/sdr]\n")
		log.Printf("[hz.tools/sdr] mmx:  %t\n", mmx)
		log.Printf("[hz.tools/sdr] avx:  %t\n", avx)
		panic("hz.tools/sdr: CPU does not support SIMD")
	}

	Backends = append(Backends, "mmx")
	Backends = append(Backends, "avx")
}

// vim: foldmethod=marker

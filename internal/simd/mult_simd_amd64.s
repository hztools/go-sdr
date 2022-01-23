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

#include "helpers_amd64.h"

// func mmxScaleComplex([4]float32, []complex64)
TEXT ·mmxScaleComplex(SB), $0-40
    MOVUPS s+0(FP), X1

    slice_size(size, 16, $8, BX)
    slice_addr(addr, 16, DX)

    MOVQ DX, SI
    ADDQ BX, SI

    // Now let's load in the size of the "stride". We process 4 32 bit (4 byte)
    // floats at a time (two complex64s).
    MOVQ $(4*4), BX

    // SI: Current pointer into the complex array
    // DX: Base complex array pointer address
    // BX:  Stride length
    // X1:  [4]float32 to multiply another [4]float32 by

    SUBQ BX, SI

scale_complex_loop:
    MOVUPS (SI), X0
    MULPS X1, X0
    MOVUPS X0, (SI)

    SUBQ BX, SI
    CMPQ SI, DX
    JGE scale_complex_loop
    RET

// vim: foldmethod=marker

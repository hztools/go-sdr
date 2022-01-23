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

// func mmxAddComplex([]complex64, []complex64, []complex64)
TEXT ·mmxAddComplex(SB), $0-72
    slice_size(a_size, 0, $8, R8)

    // Load up end of the A array
    slice_addr(a_ptr_addr, 0,  SI)
    ADDQ R8, SI

    // Load up end of the B array
    slice_addr(b_ptr_addr, 24, DI)
    ADDQ R8, DI

    // Load up the end of the C array
    slice_addr(c_ptr_addr, 48, CX)
    ADDQ R8, CX

    // MOVQ this above once it's working
    slice_addr(a_base_addr, 0,  DX)

    // Side to decrement address by.
    MOVQ $(4*4), BX

    // SI: Current A pointer
    // DI: Current B pointer
    // CX: Current C pointer
    // DX: Base A pointer address
    // BX:  Stride length

    SUBQ BX, SI
    SUBQ BX, DI
    SUBQ BX, CX

add_complex_loop:

    MOVUPS (SI), X0
    MOVUPS (DI), X1
    ADDPS X1, X0
    MOVUPS X0, (CX)

    // Decrement all three pointers
    SUBQ BX, SI
    SUBQ BX, DI
    SUBQ BX, CX
    CMPQ SI, DX
    JGE add_complex_loop
    RET


// vim: foldmethod=marker

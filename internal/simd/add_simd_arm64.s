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

// +build arm64 !sdr.nosimd

#include "helpers_arm64.h"

// func neonAddComplex(a []complex64, b []complex64, dst []complex64)
TEXT ·neonAddComplex(SB), $0-72
    slice_addr(a, 0,  R1)
    slice_addr(b, 24, R2)
    slice_addr(c, 48, R3)

    slice_size(a_size, 0, $8, R8)
    ADD R1, R8, R8

    // We drop down one since we compare after the VLD which post-increments our
    // pointer.
    MOVD $16, R7
    SUB R7, R8, R8

    // +----+----------------+
    // | R1 | a pointer      |
    // | R2 | b pointer      |
    // | R3 | c pointer      |
    // | R8 | a end address  |
    // +----+----------------+

add_complex_loop:

    VLD1.P 16(R1), [V1.S4]
    VLD1.P 16(R2), [V2.S4]

    // VFADD V1.S4, V2.S4, V1.S4
    WORD $0x4e21d441;

    VST1.P [V1.S4], 16(R3)

    CMP R1, R8
    BGE add_complex_loop

    RET

// vim: foldmethod=marker

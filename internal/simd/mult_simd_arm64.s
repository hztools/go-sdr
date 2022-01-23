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

#include "helpers_arm64.h"

// func neonScaleComplex(scaler [4]float32, a []complex64)
TEXT ·neonScaleComplex(SB), $0-40
    MOVD $scaler+0(FP), R1
    VLD1 (R1), [V2.S4]
    // VLD1R, and change scaler to a float32?

    slice_addr(a,      16,     R1)
    slice_size(a_size, 16, $8, R8)
    ADD R1, R8, R8

    // We drop down one since we compare after the VLD which post-increments our
    // pointer.
    MOVD $16, R7
    SUB R7, R8, R8

    // +----+----------------+
    // | R1 | a pointer      |
    // | V2 | scaler values  |
    // | R8 | a end address  |
    // +----+----------------+

scale_complex_loop:

    VLD1 (R1), [V1.S4]

    // VFADD V1.S4, V2.S4, V1.S4
    WORD $0x6e21dc41;

    VST1.P [V1.S4], 16(R1)

    CMP R1, R8
    BGE scale_complex_loop

    RET

// func neonRotateComplex([4]float32, [4]float32, []complex64)
TEXT ·neonRotateComplex(SB), $0-56
    MOVD $scaler_real+0(FP), R1
    VLD1 (R1), [V15.S4]

    MOVD $scaler_imag+16(FP), R1
    VLD1 (R1), [V16.S4]

    slice_addr(a,      32,     R1)
    slice_size(a_size, 32, $8, R8)
    ADD R1, R8, R8

    // We drop down one since we compare after the VLD which post-increments our
    // pointer.
    MOVD $32, R7
    SUB R7, R8, R8

    // +-----+----------------+
    // | R1  | a pointer      |
    // | V15 | real scalers   |
    // | V16 | imag scalers   |
    // | R8  | a end address  |
    // +-----+----------------+

rotate_complex_loop:

    VLD2 (R1), [V0.S4, V1.S4]

    WORD $0x6e20dde2; // VFMUL V0, V15, V2
    WORD $0x6e21de03; // VFMUL V1, V16, V3
    WORD $0x4ea3d44a; // VFSUB V3, V2, V10
    WORD $0x6e20de02; // VFMUL V0, V16, V2
    WORD $0x6e21dde3; // VFMUL V1, V15, V3
    WORD $0x4e22d46b; // VFADD V2, V3, V11

    VST2.P [V10.S4, V11.S4], 32(R1)

    CMP R1, R8
    BGE rotate_complex_loop

    RET

// vim: foldmethod=marker

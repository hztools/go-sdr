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

// +build !sdr.nosimd

#include "internal/simd/helpers_arm64.h"

#define FMINUSONE    $0xbf800000
#define FRECIP127_5  $0x3c008080

// func neonConvU8ToC64(s1 SamplesU8, s2 SamplesC64)
TEXT ·neonConvU8ToC64(SB), $0-42
    slice_addr(u8,  0,  R1)
    slice_addr(c64, 24, R2)
    slice_size(u8_len, 0, $2, R3)

    ADD R1, R3, R3
    MOVD $16, R8
    SUB R8, R3, R3

    // R1: uint8 addr
    // R2: c64 addr
    // R3: uint8 end (4 short)

    // Load 127.5 4 times into V2
    MOVW FRECIP127_5, R8
    VMOV R8, V5.S4

    MOVW FMINUSONE, R9
    VMOV R9, V4.S4

u8_to_c64_loop:

    VLD1.P 16(R1), [V0.B16]

    // Assuming USHLL{,2} SRC, DST, SHIFT

    // TODO: Use float16 to preform mult, and mult with extend to target
    // vectors

    WORD $0x2f08a410; // USHLL V0.B8 V16.H8, $0
    WORD $0x6f08a411; // USHLL2 V0.B16 V17.H8, $0

    WORD $0x2f10a600; // USHLL V16.H4, V0.S4
    WORD $0x6f10a601; // USHLL2 V16.H8, V1.S4

    WORD $0x2f10a622; // USHLL2 V17.H4, V2.S4
    WORD $0x6f10a623; // USHLL2 V17.H8, V3.S4

    WORD $0x6e21d800; // VUCVTFS V0.S4, V0.S4
    WORD $0x6e21d821; // VUCVTFS V1.S4, V1.S4
    WORD $0x6e21d842; // VUCVTFS V2.S4, V2.S4
    WORD $0x6e21d863; // VUCVTFS V3.S4, V3.S4

    // Some hacks here:
    //
    // Usually we can convert from {0x00 -> 0xFF} to {-1 -> +1} by
    // converting to a float (0 to 255), subtract 127.5 (-127.5 to 127.5)
    // and then divide by 127.5 (-1 to +1). However, doing a divide is pricy.
    //
    // So, what we're doing here instead is multiplying by 1/127.5, (0 to 2)
    // and then adding -1 (-1 to 1). This means we can use the fused multiply-add
    // (or FMLA -- VFMLA in Go terms) to preform this operation with one
    // instruction.
    //
    // Strictly speaking this is a loss of some precision (since 1/127.5 is
    // fairly small), so uh, yeah. I hope no one notices.

    VMOV V4.B16, V13.B16
    VMOV V4.B16, V14.B16
    VMOV V4.B16, V15.B16
    VMOV V4.B16, V16.B16

    WORD $0x4e25cc0d; // VFMLA V0.S4, V5.S4, V13.S4
    WORD $0x4e25cc2e; // VFMLA V1.S4, V5.S4, V14.S4
    WORD $0x4e25cc4f; // VFMLA V2.S4, V5.S4, V15.S4
    WORD $0x4e25cc70; // VFMLA V3.S4, V5.S4, V16.S4

    VST1.P [V13.S4], 16(R2)
    VST1.P [V14.S4], 16(R2)
    VST1.P [V15.S4], 16(R2)
    VST1.P [V16.S4], 16(R2)

    CMP R1, R3
    BGE u8_to_c64_loop

    RET

// vim: foldmethod=marker

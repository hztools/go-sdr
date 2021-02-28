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

#include "internal/simd/helpers_amd64.h"

// func mmxConvU8ToC64(s1 SamplesU8, s2 SamplesC64)
TEXT ·mmxConvU8ToC64(SB), $0-42
    // TODO(paultag): Check to ensure that AVX2 and AVX are supported.

    MOVQ $(0xFFFFFFFF), AX
    MOVQ AX, X1
    PMOVZXBD X1, X1
    VCVTDQ2PS X1, X1

    MOVQ $(0x02020202), AX
    MOVQ AX, X2
    PMOVZXBD X2, X2
    VCVTDQ2PS X2, X2

    DIVPS X2, X1

    /* Right, now, let's set up the base address and pointer address for the
     * uint8 array */
    slice_size(u8size, 0, $2, R8)
    slice_addr(u8,     0,     DI)
    slice_addr(u8,     0,     R9)
    ADDQ R8, DI
    MOVQ $(4*1), CX

    /* Now let's set up the pointer address for the float32 array. We'll be
     * using the uint8 to determine the size, so we don't need the base. */
    slice_size(c64size, 24, $8, R8)
    slice_addr(c64,     24,     SI)
    ADDQ R8, SI
    MOVQ $(4*4), BX

    /* +----+---------------------------+
     * | X1 | [4]float32 of "127.5"     |
     * | R9 | Base address for uint8    |
     * | DI | Address for uint8 array   |
     * | CX | DI decrement const        |
     * | SI | Address for float32 array |
     * | BX | SI decrement const        |
     * +----+---------------------------+ */

    /* Let's decrement (we start at the end, so let's grab the last 4 elements
     * first, and walk back to the head of the list */
    SUBQ BX, SI
    SUBQ CX, DI

u8_to_c64_loop:
    /* Scoop up 4 bytes (as a 8 bit integer) and throw those into xmm0 as a
     * int32. Since we're a uint8 and we're going into an int32, we don't have
     * to worry about signededness. */
    PMOVZXBD (DI), X0

    VCVTDQ2PS X0, X0 /* Convert the int32s into float32s */

    SUBPS X1, X0 /* Subtract 127.5 */
    DIVPS X1, X0 /* Divide them each by 127.5 */

    /* And finally dump those back to RAM, but in the float32 array */
    MOVUPS X0, (SI)

    SUBQ BX, SI
    SUBQ CX, DI

    CMPQ DI, R9
    JGE u8_to_c64_loop
    RET

// vim: foldmethod=marker

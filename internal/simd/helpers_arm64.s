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

// +build sdr.simdtest

#include "helpers_arm64.h"

// func InternalBufferAddr([]complex64) uintptr
TEXT ·InternalBufferAddr(SB), $0-24
    slice_addr(addr, 0, R0)
    MOVD R0, ret+24(FP)
    RET

// func InternalBufferLen([]complex64) uint
TEXT ·InternalBufferLen(SB), $0-24
    slice_len(len, 0, R0)
    MOVD R0, ret+24(FP)
    RET

// func InternalBufferSize([]complex64) uint
TEXT ·InternalBufferSize(SB), $0-24
    slice_size(size, 0, $8, R0)
    MOVD R0, ret+24(FP)
    RET

// func InternalBufferSizeSecond([]complex64) uint
TEXT ·InternalBufferSizeSecond(SB), $0-48
    slice_size(size, 24, $8, R0)
    MOVD R0, ret+48(FP)
    RET

// vim: foldmethod=marker

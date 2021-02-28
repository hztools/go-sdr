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

// +build amd64 !sdr.nosimd

// func mmxCPUSupport() bool
TEXT ·mmxCPUSupport(SB), $0
    // We're running a CPUID with a 1 in AX, which will return the
    // features the CPU supports.
    MOVQ $1, AX
    CPUID

    // Next, we're going to AND out the 23rd bit and mangle that into a
    // 1 or a 0 by shifting back by 23.
    MOVQ $0x800000, R8
    ANDQ R8, DX
    SHRQ $23, DX

    // Drop it on the stack and return.
    MOVQ DX, c1+0(FP)
    RET


// func avxCPUSupport() bool
TEXT ·avxCPUSupport(SB), $0
    // We're running a CPUID with a 1 in AX, which will return the
    // features the CPU supports.
    MOVQ $1, AX
    CPUID

    // Next, we're going to AND out the 28rd bit and mangle that into a
    // 1 or a 0 by shifting back by 28
    MOVQ $0x10000000, R8
    ANDQ R8, CX
    SHRQ $28, CX

    // Drop it on the stack and return.
    MOVQ CX, c1+0(FP)
    RET


// vim: foldmethod=marker

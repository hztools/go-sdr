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

// slice_addr will pull the address from the struct according to its ABI,
// and move that quadward into the register provided.
// 
// For instance, if there is a struct 8 bytes into the frame pointer as a
// passed argument, the following will put the starting address of the struct
// into the BX register:
//
// slice_addr(addr, 8, BX)
//
#define slice_addr(NAME, OFFSET, REGISTER) \
    MOVQ NAME+OFFSET(FP), REGISTER

// slice_len will pull the length of the struct from the struct according to
// its ABI, and move that quadward into the register provided.
// 
// For instance, if there is a struct 8 bytes into the frame pointer as a
// passed argument, the following will put the length of the struct
// into the BX register:
//
// slice_len(length, 8, BX)
//
#define slice_len(NAME, OFFSET, REGISTER)  \
    MOVQ NAME+(8+OFFSET)(FP), REGISTER

// slice_size will pull the length of the struct from the struct according to
// its ABI, multiply that by the element size (passed as a const or as a
// register), and move the resulting quadward int into the target register.
// 
// For instance, if there is a struct with 8 byte members (64 bit) starting 8
// // bytes into the frame pointer as a passed argument, the following will
// put the size of the struct into the BX register:
//
// slice_size(size, 8, $8, BX)
//
#define slice_size(NAME, OFFSET, ELEMENT_SIZE, REGISTER) \
    slice_len(NAME, OFFSET, AX) \
    MOVQ ELEMENT_SIZE, REGISTER \
    MULQ               REGISTER \
    MOVQ AX,           REGISTER

// vim: foldmethod=marker

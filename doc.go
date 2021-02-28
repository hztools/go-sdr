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

// Package sdr contains go fundemental types and helpers to allow for
// reading from and writing to software defined radios.
//
// The interfaces and functions exposed here are designed to mirror and behave
// in a way that is expected and not supprising to a Go developer. A lot of the
// design here is taken from the Go io package. A new set of interfaces are
// required in order to provide a set of tools to work with reading and writing
// IQ samples.
//
// Since conversion between IQ formats is very expensive, this package operates
// on a generic "sdr.Samples" type, which is a vector of IQ samples, in some
// format. Reading and writing time sensitive IQ data should usually be kept
// in its native format, and allow for a non-real-time conversion on a seperate
// thread to take place.
//
// Most code designed to create and consume IQ data likely wants data
// in the complex64 format, so an explicit cast or conversion should be made
// before doing signal processing.
package sdr

// vim: foldmethod=marker

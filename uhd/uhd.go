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

package uhd

// #cgo pkg-config: uhd
//
// #include <uhd.h>
import "C"

type uhdError int

var (
	// ErrInvalidDevice is returned when ...
	ErrInvalidDevice uhdError = 1

	// ErrIndex is returned when ...
	ErrIndex uhdError = 10

	// ErrKey is returned when ...
	ErrKey uhdError = 11

	// ErrNotImplemented is returned when ...
	ErrNotImplemented uhdError = 20

	// ErrUSB is returned when ...
	ErrUSB uhdError = 21

	// ErrIO is returned when ...
	ErrIO uhdError = 30

	// ErrOS is returned when ...
	ErrOS uhdError = 31

	// ErrAssertion is returned when ...
	ErrAssertion uhdError = 40

	// ErrLookup is returned when ...
	ErrLookup uhdError = 41

	// ErrType is returned when ...
	ErrType uhdError = 42

	// ErrValue is returned when ...
	ErrValue uhdError = 43

	// ErrRuntime is returned when ...
	ErrRuntime uhdError = 44

	// ErrEnvironment is returned when ...
	ErrEnvironment uhdError = 45

	// ErrSystem is returned when ...
	ErrSystem uhdError = 46

	// ErrExcept is returned when ...
	ErrExcept uhdError = 47

	// ErrBoostException is returned when ...
	ErrBoostException uhdError = 60

	// ErrStdException is returned when ...
	ErrStdException uhdError = 70

	// ErrUnknown is returned when ...
	ErrUnknown uhdError = 100
)

// Error implements the error type.
func (u uhdError) Error() string {
	switch u {
	case ErrInvalidDevice:
		return "UHD: Invalid Device"
	case ErrIndex:
		return "UHD: Index Error"
	case ErrKey:
		return "UHD: Key Error"
	case ErrNotImplemented:
		return "UHD: Not Implemented"
	case ErrUSB:
		return "UHD: USB Error"
	case ErrIO:
		return "UHD: I/O Error"
	case ErrOS:
		return "UHD: OS Error"
	case ErrAssertion:
		return "UHD: Assertion Invalid"
	case ErrLookup:
		return "UHD: Lookup Error"
	case ErrType:
		return "UHD: Type Error"
	case ErrValue:
		return "UHD: Value Error"
	case ErrRuntime:
		return "UHD: Runtime Error"
	case ErrEnvironment:
		return "UHD: Environment Error"
	case ErrSystem:
		return "UHD: System Error"
	case ErrExcept:
		return "UHD: Exception"
	case ErrBoostException:
		return "UHD: boost::Exception"
	case ErrStdException:
		return "UHD: std::Exception"
	case ErrUnknown:
		return "UHD: Unknown"
	default:
		return "UNKNOWN"
	}
}

func rvToError(err C.uhd_error) error {
	if err == 0 {
		return nil
	}
	return uhdError(err)
}

// vim: foldmethod=marker

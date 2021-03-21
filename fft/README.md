# hz.tools/sdr/fft

This package contains helpers for implementing and using FFTs within the
hz.tools ecosystem. There are many ways of doing an FFT, from OpenCL (clFFT),
to fftw, to a pure-go FFT implementation you write up for fun.

In an effort to not tie this library to a single FFT library, this defines a
generic interface to use FFTs within the codebase, and user code can pass
an FFT backend that makes sense for where it is.

Additionally, this contains helpers for working with the FFT output, such as
getting frequency widths, or bin indexes by frequency.

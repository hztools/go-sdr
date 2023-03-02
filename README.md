# hz.tools/sdr

> :warning: Please read [Expectations within this Organization](https://github.com/hztools/.github/tree/main/profile#expectations-within-this-organization) before using it.

[![Go Reference](https://pkg.go.dev/badge/hz.tools/sdr.svg)](https://pkg.go.dev/hz.tools/sdr)
[![Go Report Card](https://goreportcard.com/badge/hz.tools/sdr)](https://goreportcard.com/report/hz.tools/sdr)

Package sdr contains go fundamental types and helpers to allow for
reading from and writing to software defined radios.

The interfaces and functions exposed here are designed to mirror and behave
in a way that is expected and not surprising to a Go developer. A lot of the
design here is taken from the Go io package. A new set of interfaces are
required in order to provide a set of tools to work with reading and writing
IQ samples.

| SDR                                    | Format     | RX/TX  | State |
|----------------------------------------|------------|--------|-------|
| [rtl](rtl/README.md)                   | u8         | RX     | Good  |
| [HackRF](hackrf/README.md)             | i8         | RX/TX  | Good  |
| [PlutoSDR](pluto/README.md)            | i16        | RX/TX  | Good  |
| [rtl kerberos](rtl/kerberos/README.md) | u8         | RX     | Good  |
| [uhd](uhd/README.md)                   | i16/c64/i8 | RX/TX  | Good  |
| [airspyhf](airspyhf/README.md)         | c64        | RX     | Exp   |

## Toggles for building hz.tools/sdr.

| Build Flag     | Supported | Description                                        |
|----------------|-----------|----------------------------------------------------|
| sdr.nosimd     | yes       | Build without any SIMD ASM (useful for older CPUs) |
| sdr.nortl      | yes       | Build without any RTL-SDR support                  |
| sdr.rtl.old    | yes       | Disable new API surface support for compat         |
| sdr.nohackrf   | yes       | Build without any HackRF support                   |
| sdr.nopluto    | yes       | Build without any Pluto / iio support              |
| sdr.nouhd      | yes       | Build without any UHD support                      |
| sdr.noairspyhf | yes       | Build without any AirspyHF+ Support                |
| static         | yes       | Internally prepare for a static build              |

### -tags=static

When building a static binary, the `-tags=static` flag will pass a `--static`
flag to `pkg-config` to get all the right dependencies to build the .a rather
than the .so into the binary. However, due to the LDFLAGS restrictions, we
can't pass the other magic required to actually build a static binary yet. For
that, you'd have to invoke the build similar to:

```
$ go build --ldflags='-extldflags "-static"' -tags=static .
```

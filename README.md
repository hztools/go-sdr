# hz.tools/sdr

Package sdr contains go fundamental types and helpers to allow for
reading from and writing to software defined radios.

The interfaces and functions exposed here are designed to mirror and behave
in a way that is expected and not surprising to a Go developer. A lot of the
design here is taken from the Go io package. A new set of interfaces are
required in order to provide a set of tools to work with reading and writing
IQ samples.

| SDR                                    | Format   | RX/TX  | State |
|----------------------------------------|----------|--------|-------|
| [rtl](rtl/README.md)                   | u8       | RX     | Good  |
| [HackRF](hackrf/README.md)             | i8       | RX/TX  | Good  |
| [PlutoSDR](pluto/README.md)            | i16      | RX/TX  | Good  |
| [rtl kerberos](rtl/kerberos/README.md) | u8       | RX     | Good  |
| [lime](lime/README.md)                 | i16/c64  | RX/TX  | Exp   |
| [uhd](uhd/README.md)                   | i16/c64  | RX/TX  | Exp   |


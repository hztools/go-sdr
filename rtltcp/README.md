# RTL-TCP hz.tools/sdr driver

The `rtl_tcp` program will serve an rtl-sdr over a TCP connection, and stream
tuned IQ samples back to the client. The IQ stream is one-way, and is not
something that can be stopped, and the command stream is one-way, where errors
are not returned to the caller.

This package implements both the `sdr.Sdr` client interface, as well as a
server interface, to serve things like a HackRF or PlutoSDR via the `rtl_tcp`
protocol.

For clients, StartRx + reading from the buffer ought to be done as fast as
possible with this driver, or the windows may back up and cause problems for
the server.

| | |
|-------------|----|
| Format Type | U8 |
| Receiver    | ✓  |
| Transmitter | ✗  |


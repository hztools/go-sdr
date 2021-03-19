# Kerberos RTL-SDR hz.tools/sdr driver

The Kerberos driver is pretty unique. It's built on top of the rtl driver,
except it has 4 SDRs that are tied together to the same lock, to allow
coherent RX streams.

There's two built-in helpers, and some code to align the streams using the
on-chip RNG to sync clocks. The first will stich together 4 SDRs in adjacent
frequencies to a single sample stream at 4x the sample rate. The second will
align all 4 on the same frequency, in sample lock for coherent applications.

| | |
|-------------|----|
| Format Type | U8 |
| Receiver    | ✓  |
| Transmitter | ✗  |


# RTL-SDR E4K Tuner

The rtl-sdr can actually be a prety complex matrix of hardware and tuners. One
of the higher end rtl-sdr configurations uses the Elonics E4000 (E4K) tuner.

In addition to the baseband gain stage, this contains a "6-stage" IF gain stage.
This allows for some pretty robust control over the rx gain configuration, and
can make for a much more effective receive configuration.

It is, however, a lot more complicated and not well supported in the rtl-sdr
C library. This is a pure-go set of bindings to speak "E4K gain", to compute
and decode what the gain of a specific confguration ought to be.

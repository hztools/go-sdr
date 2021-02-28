# hz.tools/sdr

Package sdr contains go fundemental types and helpers to allow for
reading from and writing to software defined radios.

The interfaces and functions exposed here are designed to mirror and behave
in a way that is expected and not supprising to a Go developer. A lot of the
design here is taken from the Go io package. A new set of interfaces are
required in order to provide a set of tools to work with reading and writing
IQ samples.

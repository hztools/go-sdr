sdr(7) -- hz.tools sdr flag conventions
=======================================

## DESCRIPTION

Tools that use `hz.tools/cli` have a common set of SDR flags with a common
method of processing them. Tools can optionally prefix flags, if they accept
multiple SDRs (e.g., `--tx-sdr`). Flags also may be set via the env.

## OPTIONS

In addition to the flags below, some flags that are common to `hz.tools` can
be found as well (such as `--sdr` or `--frequency`). For more information
on those, please refer to the manpage I haven't written yet. Sorry about that.

  * `--sdr` (`${RF_SDR}`):
    Specificy what type of SDR to use. These can be limited at compile-time
    using the `sdr.no*` build tags, such as `sdr.nortl`, which will exclude the
    rtl driver(s). The current set of known SDRs is:

    | name | description |
    | ---- | ----------- |
    | airspyhf | Airspy HF+ Discovery (using `libairspyhf`) |
    | hackrf   | HackRF One (using `libhackrf`) |
    | pluto    | PlutoSDR (using `libiio`) |
    | rtl      | RTL-SDR (using `librtlsdr`) |
    | uhd      | UHD USRP SDR, like the Ettus B210 or Ettus B200mini (using `libuhd`) |
    
    That list is provided for convenience -- additional radios may be added or
    removed depending on compile-time configuration and specific library
    version.

  * `--frequency` (`${RF_FREQUENCY}`):
    Set the loaded SDR's center frequency to the provided frequency, provided
    in `hz.tools/rf.ParseHz` syntax. Some examples are `10Hz`, `1.3kHz` or
    `100GHz`.

  * `--sample-rate` (`${RF_SAMPLE_RATE}`):
    Set the loaded SDR's sample rate to the provided number of samples per
    second. While this is *technically* a frequency, it does not accept the
    frequency format. A sample rate of 1 Megasamples per second is specificed
    as 1000000, **not** 1MHz or 1Msps.

  * `--gains` (`${RF_GAINS}`):
    Key-value list of the SDR's Gain Stages to set. For example, using an
    RTL-SDR, you may use `--gains=Tuner=30`, or for an Ettus B210,
    `--gains=RX1PGA=20,TX0PGA=10`. Refer to documentation (or runtime) to
    enumerate supported Gain Stages. A list is provided below that may become
    out of date at <WELL KNOWN GAIN STAGES>.

## AIRSPYHF FLAGS

  * `--airspy-dsp` (`${RF_AIRSPY_DSP}`):
    Boolean flag to enable or disable the AirspyHF+ Discovery's DSP
    Processing.

  * `--airspy-serial` (`${RF_AIRSPY_SERIAL}`):
    Serial Number of the AirspyHF+ Discovery to use.

## PLUTOSDR FLAGS

  * `--pluto-uri` (`${RF_PLUTO_URI}`):
    URI of the libiio endpoint for the desired PlutoSDR. The documented `libiio`
    backends are `usb:`, `ip:`, `local:` and `usb:ip`, such as
    `ip:192.168.100.2`.

## RTL-SDR FLAGS

  * `--rtl-device-index` (`${RF_RTL_DEVICE_INDEX}`):
    Device Index of the RTL-SDR dongle to use. This is the same index as is
    used throughout the `rtl_*(1)` tools, but is dependent on the system
    usb order, and may not be stable over time.

  * `--rtl-serial` (`${RF_RTL_SERIAL}`):
    Device Serial of the RTL-SDR dongle to use. This will scan all the attached
    RTL-SDR dongles for their Serial, and select the first that matches. If
    no serial is set on the provided RTL-SDR, one can be set using
    `rtl_eeprom(1)`.

## RTLTCP FLAGS

  * `--rtltcp-host` (`${RF_RTLTCP_HOST}`):
    Network name of the host to connect to, be it a FQDN or an IP. This is
    expected to be a `rtl_tcp` server or a compatable interface (as understood
    within `https://hz.tools/rtl_tcp/`, or implemented in `hz.tools/sdr/rtltcp`).

  * `--rtltcp-port` (`${RF_RTLTCP_PORT}`):
    TCP port to connect to on the host provided by `--rtltcp-host`.

## UHD FLAGS

  * `--uhd-args` (`${RF_UHD_ARGS}`):
    Arguments to `libuhd` as defined in `uhd::stream_args_t::args`. This
    can be used to control buffer sizing or internal behavior for things like
    underflow or overflows.

  * `--uhd-buffer-length` (`${RF_UHD_BUFFER_LENGTH}`):
    *DEPRECATED*: Internal buffer length of the UHD RX path.

  * `--uhd-rx-channel` (`${RF_UHD_RX_CHANNEL}`):
    Select which RX channel to use when using `StartRx`. This is 0 indexed,
    and depends on the Radio -- for instance, the Ettus B210 has 2 channels
    (meaning: `--uhd-rx-channel=0` or `--uhd-rx-channel=1`), while the
    Ettus B200mini has only one.

  * `--uhd-rx-channels` (`${RF_UHD_RX_CHANNELS}`):
    Select which RX channels to use when using `StartCoherentRx`. This is
    the same indexing scheme as the `--uhd-rx-channel` flag.

  * `--uhd-sample-format` (`${RF_UHD_SAMPLE_FORMAT}`):
    Select the Sample Format of the `sdr.Sdr` interface. This controls
    what `sdr.Sdr.SampleFormat()` returns, as well as the returned
    `Reader` or `Writer` returned by `StartRx` or`StartTx`. This conversion
    will take place on the FPGA, and will be a lot faster than the software
    based `hz.tools/sdr/stream.ConvertReader` method, while trading off on
    I/O bandwidth. The current set of known Sample Format is:
    
  | flag | format | name |
  | ---- | ------ | ---- |
  | i8  | SampleFormatI8  | [][2]int8   |
  | i16 | SampleFormatI16 | [][2]int16  |
  | c64 | SampleFormatC64 | []complex64 |

  * `--uhd-time-source` (`${RF_UHD_TIME_SOURCE}`):
    Clock source to control timing. This is both the 1PPS trigger and
    10MHz ref. This is a string as defined by
    `uhd::usrp::multi_usrp::set_clock_source`, typical values are
    `internal` and `external`, but your specific radio may have other
    options.

  * `--uhd-tx-channel` (`${RF_UHD_TX_CHANNEL}`):
    Select which TX channel to use when using `StartTx`. This is 0 indexed,
    and depends on the Radio -- for instance, the Ettus B210 has 2 channels
    (meaning: `--uhd-tx-channel=0` or `--uhd-tx-channel=1`), while the
    Ettus B200mini has only one.

## WELL KNOWN GAIN STAGES

| radio | stage name | type | description |
| -- | -- | -- | -- |
| airspyhf | Att | RX, Attenuator | Attenuator for the RX channel |
| airspyhf | Amp | RX, Amp | Amplification for the RX channel |
| hackrf   | Amp   | TX, RX, Amp | Amplification for the RX and TX channel |
| hackrf   | RXIF  | RX, IF      | IF gain stage |
| hackrf   | RXVGA | RX, Tuner   | Tuner gain stage |
| hackrf   | TXVGA | TX, Tuner   | Tuner gain stage |
| pluto    | RX    | RX, Tuner   | Tuner gain stage |
| pluto    | TX    | TX, Tuner   | Tuner gain stage |
| uhd      | RX`N`PGA | RX, Tuner | Tuner gain for the `N`th channel, e.g. `RX0PGA` |
| uhd      | TX`N`PGA | TX, Tuner | Tuner gain for the `N`th channel, e.g. `TX0PGA` |
| rtl-sdr | Tuner | RX, Tuner | Tuner gain for the RTL-SDR |
| rtl-sdr (E4K) | IF | RX, IF | IF gain stage for the E4K RTL-SDR |

## COPYRIGHT

Copyright (c) 2023 Paul Tagliamonte <paul@k3xec.com>

## SEE ALSO

hackrf_info(1), rtl_eeprom(1), uhd_find_devices(1)

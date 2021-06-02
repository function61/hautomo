Works with at least [CC2531](https://www.zigbee2mqtt.io/information/supported_adapters.html).


Firmware flashing
-----------------

You need to flash new firmware for the CC2531.

Instructions and tool for how to flash:

- [from Linux](https://github.com/joonas-fi/cc-tool-docker)
- [from Windows](https://www.zigbee2mqtt.io/information/flashing_the_cc2531.html)

I used firmware `coordinator/Z-Stack_Home_1.2/bin/default/CC2531_DEFAULT_<date>.zip`.
(You'll find the files URL from the Linux instructions.)


Running multiple radios
-----------------------

- Our architecture supports it
- Xiaomi and IKEA don't like to mix (TODO: source)
- Increase maximum amount of Zigbee nodes (coordinator limit it pretty low)
- udev rule to make device name predictable and not mix up radios


Quickstart
----------

Make a directory for `ezhub`. If you're planning on using multiple radios (and even if you're not,
you should leave yourself a chance in the future), it's a good idea to number the directory, e.g. `ezhub1`.

Download Hautomo into that directory. It contains the `ezhub` component (as sub-command).

Install as a service:

```console
$ ./hautomo ezhub --install
```

Generate new configuration file:

```console
$ ./hautomo ezhub new-config > ezhub-config.json
```

All the values are randomized
(except [channel](https://home-assistant-guide.com/2020/10/29/choose-your-zigbee-channel-wisely/)) -
you might want to change it.

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

You should create a `udev` rule at `/etc/udev/rules.d/99-zigbee-radios.rules` to keep "COM port names"
stable across USB reconnects and multiple radios (it's a good idea for robustness even if you're not running multiple radios):

```
SUBSYSTEM=="tty", ATTRS{idVendor}=="0451", ATTRS{idProduct}=="16a8", ATTRS{serial}=="__0X00124B0018E33190", SYMLINK="ttyUSB.CC2531-ezhub1", OWNER="pi"
SUBSYSTEM=="tty", ATTRS{idVendor}=="0451", ATTRS{idProduct}=="16a8", ATTRS{serial}=="__0X00124B001CDC5B21", SYMLINK="ttyUSB.CC2531-ezhub2", OWNER="pi"
```

NOTE: replace serial numbers with your own. You can find them with:

```
$ udevadm info -a -n /dev/ttyACM0 | grep serial
```

You might want to restart to verify that device got assigned the correct path.


Quickstart
----------

Make a directory for `ezhub`. If you're planning on using multiple radios (and even if you're not,
you should leave yourself a chance in the future), it's a good idea to number the directory, e.g. `ezhub1`.

Download Hautomo into that directory. It contains the `ezhub` component (as sub-command).

Generate new configuration file:

```console
$ ./hautomo ezhub new-config > ezhub-config.json
```

All the values are randomized (except channel) -
[you might want to change it](https://home-assistant-guide.com/2020/10/29/choose-your-zigbee-channel-wisely/).

Test starting it manually:

```console
$ ./hautomo ezhub --install
```

You should get a complaint that your configuration does not match the one stored inside the radio.
Run again with `--settings-flash` to confirm the settings change. You should get this warning only
once (because you changed the network configuration).

Install as a service to start automatically on system restarts:

```console
$ ./hautomo ezhub --install
Wrote unit file to /etc/systemd/system/ezhub1.service
Run to enable on boot & to start (--)now:
	$ systemctl enable --now ezhub1
Verify successful start:
	$ systemctl status ezhub1
```


How to pair devices
-------------------

For safety, ezhub needs to be started with `--join-enable` flag. Pairing mode is enabled for only two minutes.

This is my workflow:

```console
$ systemctl stop ezhub1
$ ./hautomo ezhub --join-enable
# pair the device
# stop ezhub with Ctrl+c

# device definition was added to our state. go edit it to add its friendly name and area:
$ vim ezhub-state.json

$ systemctl start ezhub1
```

TODO: allow enabling pairing mode without stopping ezhub.


Adding support to new devices
-----------------------------

Pair your device with ezhub according to the above instructions. Pairing is lower-level and should
work even for devices we don't have specific support for yet.

### New controllable device

For devices that you can control (like light bulbs, motorized curtains etc.) that we don't have
control messages already for, I usually read the ZCL spec and add a new ZCL message struct
[like this](https://github.com/function61/hautomo/blob/d7335aba0e0acf5583af2f88e1757d32dc9c25ee/pkg/ezstack/zcl/cluster/command_local.go#L39)
and then test if
[sending that message](https://github.com/function61/hautomo/blob/d7335aba0e0acf5583af2f88e1757d32dc9c25ee/pkg/ezstack/ezhub/entrypoint.go#L267)
works.


### New sensor

If it's a sensor (it sends data to your direction), start ezhub with packet capture to capture the data it sends to us:

```console
$ ./hautomo ezhub --packet-capture=xiaomi-button-left-click.log
```

The log lines look like this:

```
2021-05-30T20:29:15.838031007+03:00 CommandType=2 Subsystem=4 Command=129 Payload=05210600547b01010054007cbc0300000301390212bc0a
```

Do something which makes the sensor send a Zigbee message (you can live tail the .log file to know
when that happens). Stop ezhub, start it back again in normal mode (if you want your network to
continue normally).

Take the payload from the packet capture and
[put it in a test](https://github.com/function61/hautomo/blob/d7335aba0e0acf5583af2f88e1757d32dc9c25ee/pkg/ezstack/ezhub/deviceadapters/xiaomibutton_test.go#L17)
that tests its transformation from Zigbee input to MQTT output.

You should be ready if the
[generic parsers](https://github.com/function61/hautomo/blob/d7335aba0e0acf5583af2f88e1757d32dc9c25ee/pkg/ezstack/ezhub/deviceadapters/genericparsers.go#L11) cover your use case.

If not, and:

- The sensor sends data according to ZCL specs
	* You can write a new [generic parser](https://github.com/function61/hautomo/commit/2550653a8491de18a1141aa8f9f91be70f010541#diff-8d5a42146d96459e82bab90022074bb2891106a2ededf42540d77363a058e284)
- The sensor uses its own data format
	* You need to write a [custom adapter](https://github.com/function61/hautomo/blob/d7335aba0e0acf5583af2f88e1757d32dc9c25ee/pkg/ezstack/ezhub/deviceadapters/xiaomibutton.go#L10)

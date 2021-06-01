ezstack = "Easy Zigbee Stack"

Goals:

- Provide the easiest to understand, high-quality & pragmatic Zigbee stack.
- Replace Zigbee2MQTT: bridge devices to be usable in Home Assistant via MQTT (& Hautomo, uses the same path).

ezstack is part of Hautomo, but is usable without Hautomo (it works as a library).


Why
---

I've been using Zigbee2MQTT for many years, but
[I've had too many issues](docs/issues-with-zigbee2mqtt.md) so I decided I want a Go-based
solution.

Other Golang Zigbee stack projects were too complex. Multi-repo projects that could've been packages
to accomplish Zigbee comms. I found them hard to understand. Or it might've just been that things
seemed complex when I didn't know about Zigbee enough. Well anyway, here we are now.
Shimmering Bee didn't have MQTT integration (I wanted to just replace Zigbee2MQTT in my setup).

I also wanted to take full control of my network by understanding the low-level tooling so when I
encounter problems I know where to look.


Current status
--------------

Supported devices: tested most Xiaomi Aqara sensors & IKEA Trådfri lights.

Works for my 30+ -node Zigbee network. I actually have two Zigbee networks to overcome the
~[20-nodes-per-CC2531 limit]((https://www.zigbee2mqtt.io/information/FAQ.html#i-read-that-zigbee2mqtt-has-a-limit-of-20-devices-when-using-a-cc2531-is-this-true))
& other issues.
So multi-Zigbee radio support can be considered to have first-class support.

I don't expect this project to be mature enough for stable use to other people very soon, because my
goal is not support as many devices as Zigbee2MQTT (it supports absolutely massive amount of devices),
but instead to support the things I need as cleanly as possible.
I'll expect to keep refactoring without having to care if I break things for someone else.

This code is public to reciprocate for the help I've received. In the rare case that you're brave
enough to run this code I'll gladly accept contributions, but will not work very hard to add
features/devices you'll need unless it benefits me as well.


Architecture
------------

Somewhat similar to [Shimmering Bee](https://shimmeringbee.io/docs/introduction/) architecture.

Basically: the CC2531 USB sticks runs as a standalone, autonomous coordinator - it is **not just a
radio** API with RX/TX frames.
Our Go-based coordinator package can therefore be thought of as the API for asking coordinator to do
things / get data from it.

We want to get data from sensors and send commands to light bulbs etc. Therefore our ezstack does its
thing by using the coordinator package, which in turn uses ZNP protocol to talk to the USB stick. The
ZNP protocol builds on top of UNP.

Application-level things like sensor data and end device commands are communicated using ZCL which
is a standardized framing structure/data format that Zigbee devices communicate with. ZCL **tries**
to standardize things like attribute IDs and values for temperature readings and for controlling lights.
Unfortunately ZCL fails to be a very good standard, and there are lots of manufacturer-specific quirks
and therefore we need abstractions to hide the warts from the user.

Logical component interaction:

```
ezstack
└── Coordinator
    ├── ZCL
    └── ZNP
        └── UNP
```

UNP = [Unified Network Processor (Interface)](https://dev.ti.com/tirex/explore/content/simplelink_cc13x2_26x2_sdk_3_10_00_53/docs/ble5stack/ble_user_guide/html/ble-stack-common/npi-index.html): Texas Instruments' protocol for communicating with their Zigbee/Bluetooth/etc radios
ZNP = [Zigbee Network Processor](http://software-dl.ti.com/simplelink/esd/plugins/simplelink_zigbee_sdk_plugin/1.60.01.09/exports/docs/zigbee_user_guide/html/zigbee/developing_zigbee_applications/znp_interface/znp_interface.html): low-level Zigbee commands used to talk UNP to Texas Instruments' Zigbee radio
ZCL = [Zigbee Cluster Library](https://zigbeealliance.org/wp-content/uploads/2019/12/07-5123-06-zigbee-cluster-library-specification.pdf): standardized message formats for features ("clusters") to turn on/off power, control lamp brightness etc.
Coordinator = [Handles network node management](https://www.zigbee2mqtt.io/information/zigbee_network.html#coordinator) (device asks to join the network), passes app-level messages to consumer (usually ezstack)
ezstack = Starts all components, receives app-level messages and provides vendor-specific parsers (sometimes ZCL is not enough, so vendors invent their own formats...) to cleaner abstractions so your app can receive "new temperature measurement from sensor XYZ"

There also exists `ezhub` ("Easy Zigbee hub") component which is a higher-level component meant to not
just interact with Zigbee messaging, but to offer abstractions for sensor devices, light bulbs etc.
and integrate with Home Assistant. High-level logical view looks like this:


ezhub
-----

```
ezhub
├── deviceadapters
├── ezstack
│   └── ...
└── homeassistantmqtt
```

`ezstack` works without `ezhub`, but `ezhub` needs `ezstack`.


Code walkthrough
----------------

Majors pieces of functionality:

- Zigbee message comes in
- MQTT client (like Home Assistant) wants to send a message to Zigbee network
- ezhub advertises its device registry to Home Assistant
- Adapter definition for a device that needs abstractions on top of non-standard 
- Incoming Zigbee message and its MQTT transformation "end-to-end" tested (for same adapter as above)


Acknowledgements
----------------

Standing on the shoulders of giants, i.e. this project wouldn't be possible without these people.

ezstack is a fork (albeit a really-major re-write) of dyrkin's
[zigbee-steward](https://github.com/dyrkin/zigbee-steward) (the project is on hiatus).
See [differences to zigbee-steward](docs/differences-to-zigbee-steward.md) for rationale of fork.

For things unclear from zigbee-steward I also learned a lot from
[Shimmering Bee's zstack](https://github.com/shimmeringbee/zstack) implementation.


Roadmap
-------

- Decouple ezstack more from Hautomo. It's mostly dependency-free, but not totally.
- Simplify dependency relationships between `ezstack/...` sub-packages. There is something hairy somewhere.
- ZCL data structures library needs more refactoring. Ideally the data structures would be generated
  from a specification (other than the ZCL spec .pdf..) instead of being manually written..
- Improve the ZCL binary serializer/deserializer. Maybe switch to
  [restruct](https://github.com/go-restruct/restruct).

Really long-term goals:

- Instead of using CC2531 to be an autonomous coordinator, I would like to use it as RX/TX radio for
  Zigbee raw packets so the radio firmware would be really simpler and we'd get much more low-level
  control. There are
  [really silly node count limitations](https://www.zigbee2mqtt.io/information/FAQ.html#i-read-that-zigbee2mqtt-has-a-limit-of-20-devices-when-using-a-cc2531-is-this-true)
  that I think is the result of the CC2531 having to keep too much state and to too much work.
  If we'd treat it as a radio, I don't think there would be any more of these silly limitations.


Alternative software
--------------------

Other Go-based Zigbee projects.

For Texas Instruments' CC2531 family of radios:

- https://godoc.org/github.com/shimmeringbee/zstack (very active project)
- https://github.com/dyrkin/zigbee-steward (on hiatus)

Other:

- https://github.com/pauleyj/gobee (for [XBee](https://en.wikipedia.org/wiki/XBee) radio)


Additional reading
------------------

- [Tasmota's Zigbee docs](https://tasmota.github.io/docs/Zigbee-Internals/) seem very understandable
  for explaining many Zigbee concepts!

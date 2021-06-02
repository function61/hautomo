Issues with Zigbee2MQTT
=======================

None of these issues alone is a show-stopper, but in aggregate I couldn't continue with Zigbee2MQTT anymore:

- I trust a statically typed language more for critical software than JavaScript
- I found it hard to debug Zigbee-level issues with Zigbee2MQTT. E.g. to this day I don't know if
  there's packet capture (or ZNP frame capture) support.
  [It looks like you need different firmware and tooling for that](https://www.zigbee2mqtt.io/how_tos/how_to_sniff_zigbee_traffic.html).
- Zigbee2MQTT security focus is not great:
	* It ships insecure configuration by default with a
	  [shared-with-everyone Zigbee network encryption key](https://github.com/Koenkk/zigbee2mqtt.io/blame/10178f159dc44ee529e7edb4d30c145520daefd9/docs/information/configuration.md#L117)
    unless you specify your own one. If you realize
	  this later and want to secure your network, changing the key requires you to re-pair all your devices!
	  Ask me how I know...
	* `permit_join` is a configurable value so it's too easy to set `permit_join: true` and forget to
	  turn it off once you've paired your devices. I've forgot to do it and it means I've left my
	  network accidentally open (security = off) and in result my neighbor paired their lightbulb to
	  my network. `permit_join` should only operate with a timer (and in fact CC2531 supports this,
	  allowing to ask to enable pairing for only a while). Having this security foot-gun already
	  hurt my security.
- It is easy to start with insecure configuration, but if you want to secure it there was a long
  time where there was no support for generating the encryption key. Because of this and that you
  had to do it manually, perhaps that lead to the my following issue about input validation for
  when I had to generate `ext_pan_id` manually.
- There was a [serious issue with lacking input validation](https://github.com/Koenkk/zigbee2mqtt/issues/5090)
  leading to hard-to-debug problem leading to my devices stopping working after restaring.
- AFAIK there are no tests for message conversions in Zigbee2MQTT, e.g.
  "this Zigbee message came in from the device -> here's how we understood it".
- Zigbee2MQTT Docker image is so big (270 MB @ 1.9.0) / has so much code it doesn't work on my
  older-generation Raspberry Pi. In comparison, the entire Hautomo image (including Zigbee hub + MQTT)
  is 32.1 MB.
- Perhaps related to the previous point about image size, Zigbee2MQTT is gaining an UI - signalling
  feature bloat IMO. UI in general is fine but does it need to be in the core, vs. making it optional
  in out-of-core component? Especially since Zigbee2MQTT due to its nature (pipe Zigbee comms to MQTT)
  has been doing fine without an UI for years.
- Zigbee2MQTT logs have occasionally "undefined" values, which I think is a symptom of untyped code
  leading to code issues.
	* Example: `(node:16) UnhandledPromiseRejectionWarning: TypeError: Cannot read property 'end' of undefined`
- The logs also have timestamps even when run as Docker container, so when you `$ docker logs zigbee2mqtt`
  you get two full timestamps for each log line.
	* Example: `2021-05-30T13:49:35.823625177Z zigbee2mqtt:info  2021-05-30 13:49:35: Disconnecting from MQTT server`
	* Docker already adds timestamps, so when running in a container I think they should be omitted.

IKEA Trådfri
============


Protocol in general
-------------------

The hub speaks [COAP](http://coap.technology/) over [DTLS](https://en.wikipedia.org/wiki/Datagram_Transport_Layer_Security).

This is a weird design choice, as COAP tries to reinvent HTTP for low-power devices. Modern devices are not so low power that you couldn't just implement HTTP. I guess DTLS is a symptom of smaller devices lacking support for current time, and regular TLS auth requires hostname-based communication which for these consumer grade devices would be hard to implement.


Locate your hub's IP address
----------------------------

You should be able to see your hub from your router's page. I didn't, because I have a Zyxel that is a piece of shit.

I found the IP address only with an NMAP scan: `$ nmap 192.168.1.0/24`


Grab & compile coap-client
--------------------------

For Raspberry Pi, follow steps in [this tutorial](https://learn.pimoroni.com/tutorial/sandyj/controlling-ikea-tradfri-lights-from-your-pi).


Create an identity for your integration
---------------------------------------

Newer versions of Trådfri firmware require you to do this step. This is taken from [here](https://github.com/ggravlingen/pytradfri/issues/90)

```
$ coap-client -m post -u Client_identity -k SECRET_IN_HUB_LABEL -e '{"9090":"IDENTITY"}' coaps://IP_OF_HUB:5684/15011/9063
```

Response of above command creates a new PSK for you.

From now on, you can make requests with `-u IDENTITY -k NEW_PSK`.


Links
-----

- https://github.com/glenndehaan/ikea-tradfri-coap-docs
- https://github.com/ggravlingen/pytradfri/issues/90
- https://github.com/jesper-lindberg/homebridge-tradfri
- https://bitsex.net/software/2017/coap-endpoints-on-ikea-tradfri/

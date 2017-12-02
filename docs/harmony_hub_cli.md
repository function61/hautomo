Guide for setting up harmonyHubCLI for research purposes.

I used tcpdump to reverse engineer how it talks to Harmony Hub,
so I could write a client in Golang.


Install
-------

```
$ docker run --rm -it ubuntu bash
$ apt update && apt install --yes nodejs npm git \
	&& git clone https://github.com/sushilks/harmonyHubCLI.git \
	&& cd harmonyHubCLI/ \
	&& npm install
```

List devices
------------

```
$ nodejs harmonyHubCli.js -l 192.168.1.153 -r devices
```


List commands for device
------------------------

```
$ nodejs harmonyHubCli.js -l 192.168.1.153 -d 'Onkyo AV Receiver' -r commands
```

Trigger command for device

```
$ nodejs harmonyHubCli.js -l 192.168.1.153 -d 'Onkyo AV Receiver' -c VolumeDown
```

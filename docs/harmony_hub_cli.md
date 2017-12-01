Guide for setting up harmonyHubCLI for research purposes.

I used tcpdump to reverse engineer how it talks to Harmony Hub,
so I could write a client in Golang.

Install
-------

```
$ docker run --rm -it ubuntu bash
$ apt update && apt install --yes vim nodejs npm git
$ git clone https://github.com/sushilks/harmonyHubCLI.git
$ cd harmonyHubCLI/
$ npm install
$ vim harmonyHubCli.js
#!/usr/bin/env nodejs

$ mv harmonyHubCli.js harmony_cli && chmod +x harmony_cli
```

List devices
------------

```
$ ./harmony_cli -l 192.168.1.153 -r devices
```


List commands for device
------------------------

```
$ ./harmony_cli -l 192.168.1.153 -d 'Onkyo AV Receiver' -r commands
```

Trigger command for device

```
$ ./harmony_cli -l 192.168.1.153 -d 'Onkyo AV Receiver' -c VolumeDown
```

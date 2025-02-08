## Usage Monitor

Track and monitor Wi-Fi usage throughout the day, logging data usage for each network accessed.


### Requirement:
- Network Manager
- Dbus
- Linux

For this program to work, you need to have **Network Manager** and **Dbus** installed.


### Installation
This repo contains the service (backend) with API endpoints 




### Build
Building requires Go

First step, clone the repo

```shell
git clone https://github.com/Clashkid155/UsageMonitor.git
```
```shell
cd UsageMonitor
go build -o usageMonitor
./usageMonitor   #To run the file
```

### Autostart
I prefer using systemd to autostart, so here's an example systemd service file.

```
[Unit]
Description=Wifi Usage Monitor
After=network.target NetworkManager.service
FailureAction=reboot

[Service]
Type=exec
Restart=always
RestartSec=1
ExecStart=/usr/bin/env php /path/to/server.php

[Install]
WantedBy=multi-user.target
```
**...To be continued**


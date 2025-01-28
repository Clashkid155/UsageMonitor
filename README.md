## Usage Monitor

Monitor your Wi-Fi usage on linux (only).

### Requirement:
- Network Manager
- Dbus

For this app to work, you need to have **Network Manager** and **Dbus** installed


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

**...To be continued**


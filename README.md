# UPSWake

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Overview

UPSWake is an application that allows you to dynamically wake servers using Wake on Lan based on the status of a NUT UPS server.

The OPA Rego language is used in order to allow for dynamic rules to be defined for when to wake a server. The status of one or many NUT UPS servers is checked against the defined rules and if the rules are met, a Wake on Lan packet is sent to the defined server.

This application is designed to be run on a Raspberry Pi or other small computer that is always on and is on the same network as the servers you wish to wake.

## Installation

### Docker

This application can be run in a Docker container. To build and run the container, run the following commands:

```bash
git clone git@github.com:TheDarthMole/UPSWake.git
cd UPSWake
docker build -t upswake .
docker run --network host -v ${PWD}/config.yaml:/config.yaml:ro -v ${PWD}/rules:/rules/:ro --name upswake upswake
```
> Note: The `--network host` flag is required to allow the container to access the host's network interface to send Wake-on-LAN packets.

Or running using pre-built image with a docker compose:

```docker-compose.yaml
version: "3.8"
services:
  upswake:
    image: thedarthmole/upswake:latest
    container_name: upswake
    network_mode: host
    command: ["serve"]
    volumes:
      - "./config.yaml:/config.yaml:ro"
      - "./rules/:/rules/:ro"
```

### Using Go

To install this application manually, you must have Go installed. Check the go.mod for the required version.

```bash
go install github.com/TheDarthMole/UPSWake@latest
```

This application can also be run manually. To do so, run the following commands:

```bash
git clone git@github.com:TheDarthMole/UPSWake.git
cd UPSWake
go build -o upswake
```

## Getting Started

A default config is created for you when you first run `upswake serve` for the first time. You will need to edit this config to suit your environment.

```yaml
nutHosts:
  - host: 192.168.1.133
    port: 3493
    name: ups1
    credentials:
      - username: upsmon
        password: bigsecret
wakeHosts:
  - name: server1
    mac: "00:00:00:00:00:00"
    broadcast: 192.168.1.255
    port: 9
    nutHost:
      name: ups1
      username: upsmon
    rules:
      - 80percentOn.rego
```

The above config allows for a flexible configuration. You can define multiple NUT hosts and multiple wake hosts. You can also define multiple rules for each wake host.

Rules are stored in the "rules" folder and are written in the OPA Rego language. The following example rule will wake the server if the UPS is on line power and the battery level is above 80%. This rule is stored in the file `80percentOn.rego`.

```rego
package authz

default allow = false

allow = true {
	input[i].Name == "cyberpower900"
	input[i].Variables[j].Name == "battery.charge"
	input[i].Variables[j].Value >= 80 # 80% or more charge
	input[i].Variables[k].Name == "ups.status"
	input[i].Variables[k].Value == "OL" # On Line (mains is present)
}
```

## Usage

```yaml
Usage:
  upsWake [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  serve       Run the UPSWake server
  wake        Manually wake a computer

Flags:
      --config string   config file (default is ./config.yaml)
  -h, --help            help for upsWake

Use "upsWake [command] --help" for more information about a command.
```

## Roadmap

- [x] Add logic to wake hosts after evaluating rules
- [x] Make serve command run continuously, add interval flag
- [ ] Add command to output all UPS json data (helps create rules)
- [ ] Bug fixes
- [ ] Add more tests
- [ ] Add more documentation
- [x] Better config validation
- [ ] Add more examples
- [ ] Change app name from UPSWake to something else
- [ ] Change rego package name from authz to something else
- [ ] Add more rego examples
- [ ] Add GitLab CI/CD to test, build and push Docker image
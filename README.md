# UPSWake

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/TheDarthMole/UPSWake)](https://goreportcard.com/report/github.com/TheDarthMole/UPSWake)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FTheDarthMole%2Fupswake.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2FTheDarthMole%2Fupswake?ref=badge_shield)

## Overview

UPSWake is an application that allows you to dynamically wake servers using Wake on Lan based on the status of 
a NUT UPS server.

The OPA Rego language is used in order to allow for dynamic rules to be defined for when to wake a server. 
The status of one or many NUT UPS servers is checked against the defined rules and if the rules are met, 
a Wake on Lan packet is sent to the defined server.

This application is designed to be run on a Raspberry Pi or other small computer that is always on and is on the same 
network as the servers you wish to wake.

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

Create a `config.yaml` file in the same directory as the application.
If a config is not provided, the application will create a default config.

```yaml
wolTargets:
  - name: server1
    mac: 12:23:45:67:89:ab
    broadcast: 192.168.1.255
    port: 9
    interval: 15s
    nutServer:
      host: 127.0.0.1
      port: 3493
      name: nut-server
      credentials:
        username: upsmon
        password: bigsecret
    rules:
      - 80percentOn.rego
```

The above config allows for a flexible configuration. 
You can define multiple NUT hosts and multiple wake hosts. 
Multiple rules can also be defined for each server to be woken.
YAML anchors can be used if the same NUT server is used for multiple servers.

> Note: the rules are evaluated in a logical OR fashion. If any of the rules are met, the host will be woken.

Rules are stored and read from the `rules` folder and are written in the OPA Rego language. 
The example rule [80percentOn.rego](./rules/80percentOn.rego) will wake the server if the UPS named "cyberpower900" is 
on line power and the battery level is above 80%.


## Usage

```yaml
Usage:
  upsWake [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  json        Retrieve JSON from a NUT server
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
- [x] Add command to output all UPS json data (helps create rules)
- [x] Bug fixes
- [x] Add more tests
- [ ] Add more documentation
- [x] Better config validation
- [ ] Add more examples
- [ ] Change app name from UPSWake to something else
- [x] Change rego package name from authz to something else
- [ ] Add more rego examples
- [x] Add GitLab CI/CD to test, build and push Docker image

## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FTheDarthMole%2Fupswake.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FTheDarthMole%2Fupswake?ref=badge_large)
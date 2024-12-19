# 🌟 UPSWake

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/TheDarthMole/UPSWake)](https://goreportcard.com/report/github.com/TheDarthMole/UPSWake)
[![Docker Image Size](https://img.shields.io/docker/image-size/thedarthmole/upswake/latest)](https://hub.docker.com/r/thedarthmole/upswake)
[![Docker Pulls](https://img.shields.io/docker/pulls/thedarthmole/upswake)](https://hub.docker.com/r/thedarthmole/upswake)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FTheDarthMole%2Fupswake.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2FTheDarthMole%2Fupswake?ref=badge_shield)

A dynamic Wake on Lan application that wakes servers based on the status of a NUT UPS server using the Rego policy language.

## 📜 Highlights

### ⚡ Efficiency

- 🌐 Lightweight API server to handle Wake on Lan requests
- 📦 Small Docker image size and low footprint
- 🥧 Can be run on a Raspberry Pi or other small computer

### 🛠️ Flexibility

- 📝 Define dynamic rules using the Rego policy language
- 📦 Multi-arch Docker image can run on any platform
- ⚙️ CLI tool to manually wake a server
- 📡 Connect to multiple NUT servers and wake multiple servers

### 🛡️ Attention to Security

<details><summary><em>Click to expand:</em> ✍️ You can verify the Docker images were built from this repository using the cosign tool.</summary>


```bash
cosign verify thedarthmole/upswake:latest \
    --certificate-identity-regexp https://github.com/TheDarthMole/upswake/ \
    --certificate-oidc-issuer https://token.actions.githubusercontent.com
```

> [!NOTE]
> This only proves that the Docker image is from this repository, assuming that no one hacks into GitHub or the repository. It does not prove that the code itself is secure.

</details>

## 🔍 Overview

UPSWake is an application that allows you to dynamically wake servers using Wake on Lan based on the status of 
a NUT UPS server.

The OPA Rego language is used in order to allow for dynamic rules to be defined for when to wake a server. 
The status of one or many NUT UPS servers is checked against the defined rules and if the rules are met, 
a Wake on Lan packet is sent to the defined server.

Upswake is designed to run on a [Raspberry Pi](https://www.raspberrypi.org/) or any small, always-on computer that
shares the same network as the servers you want to wake. 
It is ideal for environments where the servers are set to shut down using the [NUT client](https://technotim.live/posts/NUT-server-guide/) 
when the UPS switches to battery power, as Upswake provides the capability to wake them back up using intelligent rules.

## 🏎️ Getting Started

Create a `config.yaml` file in the same directory as the application.
If a config is not provided, the application will attempt to create a default config.

```yaml
nut_servers:
  - name: raspberrypi
    host: 192.168.13.37
    port: 3493
    username: upsmon
    password: bigsecret
    targets:
      - name: MyNAS
        mac: "01:23:45:67:89:01"
        broadcast: 192.168.13.255
        port: 9
        interval: 5s
        rules:
          - 80percentOn.rego
      - name: Gaming PC
        mac: "10:98:76:54:32:01"
        broadcast: 192.168.13.255
        port: 9
        interval: 15m
        rules:
          - alwaysTrue.rego
```

The above config allows for a flexible configuration where you can define multiple NUT hosts and multiple target hosts. 
Multiple rules can also be defined for each server to be woken.
YAML anchors can be used if the same NUT client is used for multiple servers.

> [!NOTE] 
> The Rego rules are evaluated in a logical OR fashion. If any of the rules evaluate to true, the host will be woken.

Rules are stored and read from the [rules](rules) folder and are written in the OPA Rego language. 
The example rule [80percentOn.rego](./rules/80percentOn.rego) will wake the server if the UPS named "cyberpower900" is 
on line power and the battery level is above 80%.

### 🐋 Deployment with Docker Compose

```yaml
version: "3.8"
services:
  upswake:
    # Choose the appropriate tag based on your need:
    # - "latest" for the latest stable version (which could become 2.x.y in the future and break things)
    # - "edge" for the latest development version running on the default branch
    # - "1" for the latest stable version whose major version is 1
    # - "1.x" for the latest stable version whose major.minor version is 1.x
    # - "1.x.y" to pin the specific version 1.x.y
    image: thedarthmole/upswake:latest
    container_name: upswake
    # Required to allow the container to access the host's network interface to send Wake-on-LAN packets
    network_mode: host
    # Restart the container automatically after reboot
    restart: always
    # Run the application as a non-root user (optional but recommended)
    # Change the user and group IDs based on your needs
    user: "1000:1000"
    # Make the container filesystem read-only (optional but recommended)
    read_only: true
    # Drop all Linux capabilities (optional but recommended)
    cap_drop: [ all ]
    # Another protection to restrict superuser privileges (optional but recommended)
    security_opt: [no-new-privileges:true]
    command: ["serve"]
    # Mount the configuration file and the rules folder as read-only volumes
    volumes:
      - "./config.yaml:/config.yaml:ro" # upswake will create a config if one doesn't exist, you may want to remove the ':ro' in that case
      - "./rules/:/rules/:ro"
```

#### 🚀 Start the application

```bash
docker compose up --detach --pull always --force-recreate
````

### ⛷️ Other Installation Methods

<details><summary><em>Click to expand:</em> 🐋 Directly run the Docker image</summary>

```bash
docker run \
  --network host \
  -v ${PWD}/config.yaml:/config.yaml:ro \
  -v ${PWD}/rules:/rules/:ro \
  --name upswake \
  thedarthmole/upswake:latest
```

> Note: The `--network host` flag is required to allow the container to access the host's network interface to send Wake-on-LAN packets.

</details>

<details><summary><em>Click to expand:</em> 🧬 Directly install upswake from its source</summary>

You need the [Go tool](https://golang.org/doc/install) to run upswake from its source.

```bash
go install github.com/TheDarthMole/UPSWake@latest
```

</details>

<details><summary><em>Click to expand:</em> 🏗️ Build upswake from its source</summary>

You need the [Go tool](https://golang.org/doc/install) to build upswake from its source.

```bash
git clone git@github.com:TheDarthMole/upswake.git
cd upswake
go build -o upswake ./cmd/upswake
```

</details>

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
- [x] Change rego package name from authz to something else
- [ ] Add more rego examples
- [x] Add GitLab CI/CD to test, build and push Docker image

## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FTheDarthMole%2Fupswake.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FTheDarthMole%2Fupswake?ref=badge_large)

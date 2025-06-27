<p>
  <a href="https://github.com/jonas-grgt/ktea/releases"><img src="https://img.shields.io/github/release/jonas-grgt/ktea.svg" alt="Latest Release"></a>
  <a href="https://github.com/jonas-grgt/ktea/actions"><img src="https://github.com/jonas-grgt/ktea/actions/workflows/ci.yml/badge.svg?branch=main" alt="Build Status"></a>
</p>

# 🫖 ktea - kafka terminal client

ktea is a tool designed to simplify and accelerate interactions with Kafka clusters.

![demo.gif](demo.gif)

## Installation

### Mac

```sh
brew tap jonas-grgt/ktea
brew install ktea
```

### Linux

Binaries available at the release page.

### Windows

Binaries available at the release page.

## Usage

### Configuration

All configuration is stored in `~/.config/ktea/config.conf`

### Cluster Management

Multiple clusters can be added.
Upon startup when no cluster is configured you will be prompted
to add one.

#### Supported Auth Methods

- No Auth
- SASL (SSL)
    - PLAIN

### Switching Tabs

To switch between tabs the meta key is required, which in most terminals needs to be enabled and will map to `Alt`.

- iterm: https://iterm2.com/faq.html
- kitty: https://sw.kovidgoyal.net/kitty/conf/#opt-kitty.macos_option_as_alt
- Mac Terminal: https://superuser.com/questions/496090/how-to-use-alt-commands-in-a-terminal-on-os-x

## Features

- *Multi-Cluster Support*: Seamlessly connect to multiple Kafka clusters and switch between them with ease.
- *Topic Management*: List, create, delete, and modify topics, including partition and offset details.
- *Record Consumption*: Consume records in text, JSON, and **Avro** formats, with powerful search capabilities.
- *Consumer Group Insights*: Monitor consumer groups, view their members, and track offsets.
- *Schema Registry Integration*: Browse, view, and register schemas effortlessly.

## Todo

- Add more authentication methods
- Add support for more message formats such as protobuf.
- Add ACL management.
- File based import/export of topics.
- Add ability to delete specific schema versions.
- Add consumption templating support.
- Many more, just file an issue requesting a feature!

## Development

### Dev cluster setup

A docker-compose setup is provided to quickly spin up a local Kafka cluster with pre-created topics, consumer groups,
commited offsets etc ...

```sh
cd docker
docker-compose up -d
```

### Generate data

After the local cluster is up and running, you can generate some data to work with, 
using`go run -tags prd ./cmd/generate`.
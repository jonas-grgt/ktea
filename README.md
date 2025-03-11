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
Coming soon

## Usage

### Switching Tabs

To switch between tabs the meta key is required, which in most terminals needs to be enabled and will map to `Alt`.

- iterm: https://iterm2.com/faq.html
- kitty: https://sw.kovidgoyal.net/kitty/conf/#opt-kitty.macos_option_as_alt
- Mac Terminal: https://superuser.com/questions/496090/how-to-use-alt-commands-in-a-terminal-on-os-x

## Features
- *Multi-Cluster Support*: Seamlessly connect to multiple Kafka clusters and switch between them with ease.
- *Topic Management*: List, create, delete, and modify topics, including partition and offset details.
- *Record Consumption*: Consume records in text, JSON, and Avro formats, with powerful search capabilities.
- *Consumer Group Insights*: Monitor consumer groups, view their members, and track offsets.
- *Schema Registry Integration*: Browse, view, and register schemas effortlessly.

## Todo
- Add support for more message formats such as protobuf.
- Add ACL management.
- File based import/export of topics.
- Add ability to delete specific schema versions.
- Add consumption templating support.
- Add sorting Topic capabilities- 
- Many more, just file an issue requesting a feature!
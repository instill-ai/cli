# CLI for Instill Core

[![Tests](https://github.com/instill-ai/cli/actions/workflows/go.yml/badge.svg?branch=main&event=push)](https://github.com/instill-ai/cli/actions/workflows/go.yml)
[![GitHub commits since latest release (by SemVer including pre-releases)](https://img.shields.io/github/release/instill-ai/cli.svg?include_prereleases&label=Release&color=lightblue)](https://github.com/instill-ai/cli/releases/latest)
[![License](https://img.shields.io/github/license/instill-ai/cli.svg?color=lightblue&label=License)](./License.md)

`inst` is a command line tool for **[Instill Core](https://github.com/instill-ai/instill-core).

## Prerequisites

- **macOS or Linux** - `inst` works on macOS or Linux, but does not support Windows yet.


## Installation

`inst` is available via [Homebrew](https://brew.sh) and as a downloadable binary from the [releases page](https://github.com/instill-ai/cli/releases).

To install:

```shell
brew install instill-ai/tap/inst
```

To uninstall:

```shell
brew uninstall inst
```

To upgrade:

```shell
brew upgrade inst
```

## Usage examples

```shell
# Check all available commands
inst help

# Deploy a local Instill Core
inst local deploy

# Undeploy a local Instill Core
inst local undeploy

# Authorization for an instance (default to managed Instill Core https://api.instill-ai.com)
inst auth login

# REST API request
inst api vdp/alpha1/pipelines
```

## Documentation

📔 **Documentation**

 Please check out the [documentation](https://www.instill.tech/docs?utm_source=github&utm_medium=banner&utm_campaign=vdp_readme) website.

📘 **API Reference**

The gRPC protocols in [protobufs](https://github.com/instill-ai/protobufs) provide the single source of truth for the VDP APIs. The genuine protobuf documentation can be found in our [Buf Scheme Registry (BSR)](https://buf.build/instill-ai/protobufs).

For the OpenAPI documentation, access http://localhost:3001 after `make all`, or simply run `make doc`.

## Contributing

Please refer to the [Contributing Guidelines](./.github/CONTRIBUTING.md) for more details.

## Community support

Please refer to the [community](https://github.com/instill-ai/community) repository.

## License

See the [LICENSE](./LICENSE) file for licensing information.

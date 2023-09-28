<h1 align="center">
  <img src="https://raw.githubusercontent.com/instill-ai/.github/main/img/cli.png" alt="Instill AI - Unstructured Data ETL Made for All" />
</h1>

<h4 align="center">
    <a href="https://www.instill.tech/?utm_source=github&utm_medium=banner&utm_campaign=cli_readme">Website</a> |
    <a href="https://discord.gg/sevxWsqpGh">Community</a> |
    <a href="https://blog.instill.tech/?utm_source=github&utm_medium=banner&utm_campaign=cli_readme">Blog</a><br/><br/>
    <a href="https://www.instill.tech/docs/?utm_source=github&utm_medium=banner&utm_campaign=cli_readme">User Manual</a> |
    <a href="https://discord.gg/sevxWsqpGh">API Reference</a><br/><br/>
    <a href="https://www.instill.tech/get-access/?utm_source=github&utm_medium=banner&utm_campaign=cli_readme"><strong>Get Early Access</strong></a>
</h4>

---

[![Tests](https://github.com/instill-ai/cli/actions/workflows/go.yml/badge.svg?branch=main&event=push)](https://github.com/instill-ai/cli/actions/workflows/go.yml)
[![GitHub commits since latest release (by SemVer including pre-releases)](https://img.shields.io/github/release/instill-ai/cli.svg?include_prereleases&label=Release&color=lightblue)](https://github.com/instill-ai/cli/releases/latest)
[![License](https://img.shields.io/github/license/instill-ai/cli.svg?color=lightblue&label=License)](./License.md)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-lightblue.svg)](.github/code_of_conduct.md)

`instill` is the command line interface for **Instill Cloud**.

## Table of contents <!-- omit in toc -->
- [What is Instill Cloud?](#what-is-instill-cloud)
- [Installation](#installation)
- [Issues and discussions](#issues-and-discussions)

## What is Instill Cloud?

Instill AI is on a mission to make AI accessible to everyone. With **Instill Cloud**, one can easily build up a data pipeline to transform unstructured data to meaningful data representations, starting to tap on the value of unstructured data.

**Instill Cloud** provides a **Pipeline** consisting of **Source Connector**, **Model**, **Logic Operator** and **Destination Connector** components, to process unstructured data to their meaningful data representations.

## Installation

You can download binaries from the [releases page][].

### Linux <!-- omit in toc -->

```
curl -Ls https://github.com/instill-ai/cli/releases/download/v0.2.0-alpha/instill_Linux_x86_64.tar.gz | tar -xzvf -
./bin/instill
```

#### MacOS <!-- omit in toc -->

To install:
```
brew install instill-ai/tap/instill
```

To upgrade:
```
brew upgrade instill-ai/tap/instill
```

## Documentation

[Instill CLI](doc/instill.1.md):
- [api](doc/instill-api.1.md)
- [auth](doc/instill-auth.1.md)
  - [login](doc/instill-auth-login.1.md)
  - [logout](doc/instill-auth-logout.1.md)
  - [status](doc/instill-auth-status.1.md)
- [instances](doc/instill-instances.1.md)
  - [add](doc/instill-instances-add.1.md)
  - [edit](doc/instill-instances-edit.1.md)
  - [list](doc/instill-instances-list.1.md)
  - [set-default](doc/instill-instances-set-default.1.md)
- [config](doc/instill-config.1.md)
  - [get](doc/instill-config-get.1.md)
  - [set](doc/instill-config-set.1.md)

```
Access Instill services from the command line.

USAGE
  instill <command> <subcommand> [flags]

CORE COMMANDS
  api:        Make an authenticated Instill API request
  auth:       Login and logout
  completion: Generate shell completion scripts
  config:     Manage configuration for instill
  help:       Help about any command
  instances:  Instances management
  local:      Local Instill Core instance

FLAGS
  --help      Show help for command
  --version   Show instill version

EXAMPLES
  $ instill api pipelines
  $ instill config get editor
  $ instill auth login

ENVIRONMENT VARIABLES
  See 'instill help environment' for the list of supported environment variables.

LEARN MORE
  Use 'instill <command> <subcommand> --help' for more information about a command.
  Read the manual at https://docs.instill.tech

FEEDBACK
  Please open an issue on https://github.com/instill-ai/cli.
```

## Usage examples

```bash
# log in
$ instill auth login

# list pipelines
$ instill api pipelines

# list models
$ instill api models

# add parameters to a GET request
$ instill api models?visibility=public

# list instances
$ instill instances list

# selected a default instance
$ instill instances set-default my-host.com

# add an instance
$ instill instances add instill.localhost \
    --oauth2 auth.instill.tech \
    --audience https://instill.tech \
    --issuer https://auth.instill.tech/ \
    --secret YOUR_SECRET \
    --client-id CLIENT_ID

# add parameters to a POST request
$ instill api -X POST oauth2/token?audience=...&grant_type=...

# add nested JSON body to a POST request
$ jq -n '{"contents":[{"url": "https://artifacts.instill.tech/dog.jpg"}]}' | instill api demo/tasks/classification/outputs --input -

# set a custom HTTP header
$ instill api -H 'Authorization: Basic mytoken' ...
```

## Issues and discussions
Please directly report any issues in [Issues](https://github.com/instill-ai/cli/issues) or [Pull requests](https://github.com/instill-ai/cli/pulls), or raise a topic in [Discussions](https://github.com/instill-ai/cli/discussions).

[Homebrew]: https://brew.sh
[releases page]: https://github.com/instill-ai/cli/releases/latest

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

`inst` is the command line interface for **Instill Core / Cloud**.

## Table of contents <!-- omit in toc -->
- [What is Instill Cloud?](#what-is-instill-cloud)
- [Installation](#installation)
- [Usage examples](#usage-examples)
- [Issues and discussions](#issues-and-discussions)

## What is Instill Cloud?

Instill AI is on a mission to make AI accessible to everyone. With **Instill Cloud**, one can easily build up a data pipeline to transform unstructured data to meaningful data representations, starting to tap on the value of unstructured data.

**Instill Cloud** provides a **Pipeline** consisting of **Source Connector**, **Model**, **Logic Operator** and **Destination Connector** components, to process unstructured data to their meaningful data representations.

## Installation

### macOS <!-- omit in toc -->

`instill` is available via [Homebrew][] and as a downloadable binary from the [releases page][].

#### Homebrew <!-- omit in toc -->

To install:
```
brew install instill-ai/tap/inst
```

To upgrade:
```
brew upgrade instill-ai/tap/inst
```

## Usage examples

```bash
# log in
$ inst auth login

# list pipelines
$ inst api pipelines

# list models
$ inst api model/models

# add parameters to a GET request
$ inst api model/models?visibility=public

# list instances
$ inst instances list

# add nested JSON body to a POST request
$ jq -n '{"contents":[{"url": "https://artifacts.instill.tech/dog.jpg"}]}' | inst api demo/tasks/classification/outputs --input -

# set a custom HTTP header
$ inst api -H 'Authorization: Basic mytoken' ...
```

## Issues and discussions
Please directly report any issues in [Issues](https://github.com/instill-ai/cli/issues) or [Pull requests](https://github.com/instill-ai/cli/pulls), or raise a topic in [Discussions](https://github.com/instill-ai/cli/discussions).

[Homebrew]: https://brew.sh
[releases page]: https://github.com/instill-ai/cli/releases/latest

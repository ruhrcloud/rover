# rover

[![Release](https://img.shields.io/github/v/release/ruhrcloud/rover)](https://github.com/ruhrcloud/rover/releases/latest)
![Go Version](https://img.shields.io/github/go-mod/go-version/ruhrcloud/rover/main?label=Go)
[![Build](https://github.com/ruhrcloud/rover/actions/workflows/build.yml/badge.svg)](https://github.com/ruhrcloud/rover/actions/workflows/build.yml)

`rover` was created to bridge the gap for printers that don’t support WebDAV by fetching scan emails via IMAP and uploading the attachments directly to a WebDAV instance.
This is especially useful since some cloud storage solutions like Nextcloud are built around WebDAV and by design do not support protocols like SMB or FTP.

## Installation
You can download pre-built binaries for Linux on the [releases](https://github.com/ruhrcloud/rover/releases) page.

If you prefer to build locally and have Go installed, you can get the latest version of `rover` by running:
```
go install github.com/ruhrcloud/rover/cmd/rover@latest
```

## Getting started

## Legal

Copyright 2025 ruhrcloud UG (haftungsbeschränkt) (<a href="mailto:legal&amp;#64;ruhrcloud.eu">legal&#64;ruhrcloud.eu</a>)<br>
SPDX-License-Identifier: MIT

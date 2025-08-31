# rover

[![Release](https://img.shields.io/github/v/release/ruhrcloud/rover?include_prereleases)](https://github.com/ruhrcloud/rover/releases/latest)
![Go Version](https://img.shields.io/github/go-mod/go-version/ruhrcloud/rover/main?label=Go)
[![CI](https://github.com/ruhrcloud/rover/actions/workflows/ci.yml/badge.svg)](https://github.com/ruhrcloud/rover/actions/workflows/ci.yml)

`rover` was created to bridge the gap for printers that don’t support WebDAV by fetching scan emails via IMAP and uploading the attachments directly to a WebDAV instance.
This is especially useful since some cloud storage solutions like Nextcloud are built around WebDAV and by design do not support protocols like SMB or FTP.

## Installation
You can download pre-built binaries for Linux on the [releases](https://github.com/ruhrcloud/rover/releases) page.

If you prefer to build locally and have Go installed, you can get the latest version of `rover` by running:
```
go install github.com/ruhrcloud/rover/cmd/rover@latest
```

## Configuration
```yaml
debug: false
tasks:
  - name: "johndoe"
    from:
      host: "mail.example.com"
      user: "johndoe@example.com"
      pass: "password"
      mailbox: "INBOX"
    to:
      base_url: "https://cloud.example.com/remote.php/webdav/johndoe"
      user: "johndoe"
      pass: "password"
    path: "/Attachments"
    tags: ["personal", "mail"]
    filter:
      recipients: ["johndoe@example.com"]
      extensions: ["pdf", "csv"]
      seen: false
    interval: "1m"
    format: "{{ slug .Subject }}{{ .OrigExt }}"
    mark_seen: true
```

## Usage

```
Usage of rover:
  -config string
        path to the config file (default "rover.yml")
```

## Legal

Copyright 2025 ruhrcloud UG (haftungsbeschränkt) (<a href="mailto:legal&amp;#64;ruhrcloud.eu">legal&#64;ruhrcloud.eu</a>)<br>
SPDX-License-Identifier: MIT

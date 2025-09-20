# rover

[![Release](https://img.shields.io/github/v/release/ruhrcloud/rover?include_prereleases)](https://github.com/ruhrcloud/rover/releases/latest)
![Go Version](https://img.shields.io/github/go-mod/go-version/ruhrcloud/rover/main?label=Go)
[![CI](https://github.com/ruhrcloud/rover/actions/workflows/ci.yml/badge.svg)](https://github.com/ruhrcloud/rover/actions/workflows/ci.yml)

`rover` was created to bridge the gap for printers that don’t support WebDAV by fetching scan emails via IMAP and uploading the attachments directly to a WebDAV instance.
This is especially useful since some cloud storage solutions like Nextcloud are built around WebDAV and by design do not support protocols like SMB or FTP.

## Installation

You can download pre-built binaries for Linux and macOS on the [releases](https://github.com/ruhrcloud/rover/releases) page.

If you prefer to build locally and have Go installed, you can get the latest version of `rover` by running:

```
go install github.com/ruhrcloud/rover/cmd/rover@latest
```

## Usage

```
Usage of rover:
  -config string
    Path to the config file (default "rover.yml").

  -verbose
    Also show connection information. Useful for debugging.
```

With the below [configuration](https://github.com/ruhrcloud/rover/blob/main/rover.yml) running `rover` would look like this:

```
# rover
2025/09/19 07:42:07 info [johndoe] starting task to run every 1m
2025/09/19 07:43:07 info [johndoe] no messages matched criteria
2025/09/19 07:43:07 info [johndoe] found 0 and uploaded 0 attachments
[...]
2025/09/19 08:12:07 info [johndoe] processed 1 and uploaded 1 attachment
```

## Configuration

This is an example configuration. Note that you can define multiple tasks to run concurrently.
For more examples have a look in the [examples](https://github.com/ruhrcloud/rover/tree/main/examples) directory in this repository.

```yaml
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

## Legal

Copyright 2025 ruhrcloud UG (haftungsbeschränkt) (<a href="mailto:legal&amp;#64;ruhrcloud.eu">legal&#64;ruhrcloud.eu</a>)<br>
SPDX-License-Identifier: MIT

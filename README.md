# BookShift

BookShift transfers e-books from network shares or email to a local folder (e.g., your Kobo device). It’s designed for simple, repeatable syncs from NAS or inbox to your reader.

## Features

- Download book files from NFS shares
- Download book files from SMB shares
- Download book attachments from an email account (IMAP)

## Usage

- Run with a config file:

  - `bookshift run -c config.yaml`
  - Or via env: `BOOKSHIFT_CONFIG_FILE=/path/to/config.yaml bookshift run`

- Useful flags for `run`:

  - `--dry-run` Show what would be done without creating, writing, or deleting files.
  - `--no-progress` Disable progress bars during downloads.

- Global:
  - `-c, --config-file` Path to configuration file (defaults to `config.yaml`).
  - `-v, --version` Print version and exit.

Notes:

- When new books are downloaded and a Kobo device is detected, the library is refreshed automatically (via NickelDBus or simulated USB plug).
- File extension matching is case-insensitive.
- Concurrency: sources are processed in parallel. Control with `concurrency` in the config (default: 3). BookShift also supports cancellation: press Ctrl+C to stop; in-flight operations will respect per-source timeouts or cancel between files/messages.

## Configuration

Top-level keys:

- `log_level`: one of `debug`, `info`, `warn`, `error` (default: `info`).
- `target_folder`: required, local folder to place downloaded books.
- `overwrite_existing_files`: overwrite existing destination files when true.
- `valid_extensions`: list of allowed extensions (e.g. `[".epub", ".kepub"]`).
- `sources`: list of source definitions; each has a `type` and a `config` block.
- `concurrency`: optional, number of sources to process in parallel (default 3).

Example config:

```yaml
log_level: info
target_folder: /mnt/books
overwrite_existing_files: false
valid_extensions: [".epub", ".kepub"]

sources:
  - type: smb
    config:
      host: fileserver.local
      port: 445 # optional (default 445)
      username: alice
      password: secret
      domain: WORKGROUP
      share: books
      folder: incoming
      keep_folderstructure: true
      remove_files_after_download: false
      timeout_seconds: 120 # optional per-source timeout

  - type: nfs
    config:
      host: nas.local
      port: 2049 # optional (default 2049)
      folder: /export/books
      keep_folderstructure: false
      remove_files_after_download: false
      timeout_seconds: 120

  - type: imap
    config:
      host: mail.example
      port: 143 # optional (default 143)
      username: reader
      password: secret
      mailbox: INBOX
      filter_field: subject # one of: to, subject
      filter_value: "[BOOK]"
      process_read_emails: false
      remove_emails_after_download: true
      timeout_seconds: 180
```

Source notes:

- SMB: `share` is the share name; `folder` is the path inside the share.
- NFS: `folder` is the exported path; remote paths use forward slashes.
- IMAP: Attachments are filtered by extension and decoded using the part’s transfer-encoding (base64, quoted-printable, etc.). When unspecified, base64 is assumed for attachments.
  Cancellation: IMAP sync checks for cancellation between messages; per-source `timeout_seconds` bounds the session.

Cancellation and timeouts:

- Press Ctrl+C to cancel an ongoing run. BookShift will stop starting new source syncs and each active source will exit promptly.
- Per-source `timeout_seconds` puts an upper bound on a source run. If exceeded, that source aborts with a timeout error and other sources continue.

## Kobo setup

### Prerequisites

BookShift integrates with [NickelMenu](https://pgaskin.net/NickelMenu/) and [NickelDBus](https://github.com/shermp/NickelDBus). Even though they are not required, it benefits from having them installed.

### Installation

Grab the `KoboRoot.tgz` file from the [latest release](https://github.com/bjw-s-labs/bookshift/releases/latest) and transfer this to your Kobo device by connecting it to your computer and placing the file in the `.kobo` folder on the exposed drive.

Finally, disconnect the device from your computer and wait for it to restart. Once the restart is complete BookShift will be installed.

### Configuration

Modify the `config.yaml` file in the `.adds/bookshift` folder to set up BookShift.

### Uninstall

Place a file named `UNINSTALL` in the `.adds/bookshift` folder to uninstall BookShift from your Kobo reader.

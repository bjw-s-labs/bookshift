# BookShift

This tool can be used to easily transfer ebooks from one location to another.
I personally use this to transfer books from my NAS to my Kobo device.

## Features

- Download book files from NFS shares
- Download book files from SMB shares
- Download book attachments from an email account

## Kobo setup

### Prerequisites

BookShift integrates with [NickelMenu](https://pgaskin.net/NickelMenu/) and [NickelDBus](https://github.com/shermp/NickelDBus).
Even though they are not required, it benefits from having them installed.

### Installation

Grab the `KoboRoot.tgz` file from the [latest release](https://github.com/bjw-s-labs/bookshift/releases/latest) and transfer this to your Kobo device by connecting it to your computer and placing the file in the `.kobo` folder on the exposed drive.

Finally, disconnect the device from your computer and wait for it to restart. Once the restart is complete BookShift will be installed.

### Configuration

Modify the `config.yaml` file in the `.adds/bookshift` folder to set up BookShift.

### Uninstall

Place a file named `UNINSTALL` in the `.adds/bookshift` folder to uninstall BookShift from your Kobo reader.

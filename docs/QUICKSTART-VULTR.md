# Quickstart on Vultr

This guide is a step-by-step on how to run SkyEye on [Vultr](https://www.vultr.com/).

It is assumed that you have set up an account, SSH keys and a billing method.

## Create a SkyEye Server

Go to https://my.vultr.com. Click "Deploy +" in the top right corner, then "Deploy New Server".

Type: Dedicated CPU

Location: Choose the location closest to the SRS server.

Plan: Select one with at least 2 vCPU and 4GB of memory. The CPU Optimized plans are ideal. Storage capacity does not matter.

Click "Step 2: Configure Software & Deploy Instance"

Operating System: If you have no particular preference, choose the most recent version of Ubuntu. Arch, CentOS, Debian, Fedora and Rocky all probably work too, if you prefer. Do not select Alpine, since [it uses a different OS library than the one used to test SkyEye](https://wiki.musl-libc.org/functional-differences-from-glibc.html).

SSH Keys: Select your SSH key.

Server 1 hostname: "skyeye-<name of dcs server here>"

Server 1 Label: "skyeye"

Deselect the "Automatic Backups" feature. SkyEye does not retain any data that needs to be backed up.

Select the "Cloud-Init User-Data" feature. Copy the contents of [`cloud-config.yaml`](../init/cloud-init/cloud-config.yaml) into a text editor.

Find the line that contains `/etc/skyeye/config.yaml`, then below it, the block under `content:`. This indented block is your SkyEye config file. Reference the [example config file](../config.yaml) and set the values as required. Remember to preserve the indentation.

Find the line that contains `ghcr.io/dharmab/skyeye:latest`. This default value will install the latest version of SkyEye **at the time the server is created**. If you want to install a specific version, replace `latest` with a version number. Example: `ghcr.io/dharmab/skyeye:v0.14.0`.

Copy the entire contents of the customized `cloud-config.yaml` file and paste it into the "User Data" box. You might also want to save this customized file for future use.

Click "Deploy".

If the configuration was correct, SkyEye should connect to your SRS server within a few minutes and announce itself with a SUNRISE broadcast. If you're comfortable with Linux, SSH into the server and check the service and logs with `systemctl status skyeye` and `journalctl -u skyeye` for any weird warnings or errors. Try some basic SkyEye commands such as a [RADIO CHECK](PLAYER.md#radio-check) and a [PICTURE](PLAYER.md#picture). Make sure the results you hear match what you see in the DCS F10 map.

## Reducing the Bill

You pay for the SkyEye server on an hourly basis. You can delete the server when you're not playing DCS to reduce your bill. Note that it's not enough to power off the server; you must delete it. 

You can recreate the server at any time by following the steps above; if you saved the customized `cloud-config.yaml` file, you can recreate the server in a few minutes. If you're an advanced user, see the [autoscaling documentation](ADMIN.md#autoscaling-experimental) for a way to automate this task.

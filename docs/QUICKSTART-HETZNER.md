# Quickstart on Hetzner Cloud

This guide is a step-by-step on how to run SkyEye on [Hetzner Cloud](https://www.hetzner.com/cloud) using local speech recogntion.

It is assumed that you have set up an account and a billing method.

## Getting Help

See [the admin guide](ADMIN.md#getting-help) for how to get help if you have a problem.

## Create a Firewall

You'll want to create a Firewall to allow the necessary network traffic for SkyEye. You only need one Firewall for all of your SkyEye servers.

Go to https://console.hetzner.cloud. Create a new project, or select an existing one.

From the project dashboard, click "Create Resource", then "Firewalls".

Create the following rules:

`INBOUND`

|Description|Protocol|Port|Port Range|Comment
|-|-|-|-|-|
|SSH|TCP|22|||
|ICMP|ICMP||||

`OUTBOUND`

|Description|Protocol|Port|Port Range|Comment
|-|-|-|-|-|
|SimpleRadio-Standalone Data|TCP|5002||Change if you use a non-standard SRS port.|
|SimpleRadio-Standalone Audio|TCP|5002||Change if you use a non-standard SRS port.|
|TacView Real-Time Telemetry|TCP|42674||Change if you use a non-standard TacView port.|
|HTTPS|TCP|443|||
|NTP|UDP|123|||

Under "Apply to", select "Select Resource", then "Label". Type the new label "skyeye" into the box and click "Add Label Selction".

Under "Name", type "skyeye".

Click "Create Firewall".

## Create a SkyEye Server

Go to https://console.hetzner.cloud. Select the project you created the Firewall in during the previous step.

From the project dashboard, click "Create Resource", then "Server".

Location: As of current writing, the best instance types for SkyEye are only available in certain regions. Check [this page](https://www.hetzner.com/cloud/) for the Locations where Dedicated vCPU CCX23 instances are available. Choose one of those locations.

Image: If you have no particular preference, choose Ubuntu. Fedora, Debian, CentOS and Rocky all probably work too, if you prefer.

Type: Choose "Dedicated vCPU" and then "CCX23".

Networking: Choose "Public IPv4" and "Public IPv6". Leave "Private network" unchecked.

**DO NOT USE A WEAK PASSWORD IN THE NEXT STEP!**

Password/SSH keys: Provide a password or SSH key. SSH keys are preferred. See [this tutorial](https://community.hetzner.com/tutorials/howto-ssh-key) for more information. You'll need to use the password or SSH key to log in to the server to look at logs or restart SkyEye, so make sure you have it saved somewhere. **DO NOT USE A SIMPLE PASSWORD** as your server is immediately going to be scanned and brute-forced by fleets of Russian, Chinese and North Korean botnets. If you use a password you _must_ use a long random password! **I CANNOT STRESS THIS ENOUGH!**

**DO NOT USE A WEAK PASSWORD IN THE PREVIOUS STEP!**

Volumes: Do not create any additional volumes.

Firewalls: Select the "skyeye" firewall you created in the previous step.

Backups: Do not enable backups. SkyEye does not retain any data that needs to be backed up.

Placement groups: Do not create any placement groups. SkyEye is a single-instance application and does not benefit from placement groups.

Labels: Add the label "skyeye" to the labels.

Cloud config:

Copy the contents of [`cloud-config.yaml`](../init/cloud-init/cloud-config.yaml) into a text editor.

Find the line that contains `/etc/skyeye/config.yaml`, then below it, the block under `content:`. This indented block is your SkyEye config file. Reference the [example config file](../config.yaml) and set the values as required. Remember to preserve the indentation.

Find the line that contains `ghcr.io/dharmab/skyeye:latest`. This default value will install the latest version of SkyEye **at the time the server is created**. If you want to install a specific version, replace `latest` with a version number. Example: `ghcr.io/dharmab/skyeye:v0.14.0`.

Copy the entire contents of the customized `cloud-config.yaml` file and paste it into the "Cloud config" box. You might also want to save this customized file for future use.

Set "Name" to something descriptive, like "skyeye-<dcs-server-name>".

Click "Create & Buy Now". 

If the configuration was correct, SkyEye should connect to your SRS server within a few minutes and announce itself with a SUNRISE broadcast. If you're comfortable with Linux, SSH into the server and check the service and logs with `systemctl status skyeye` and `journalctl -u skyeye` for any weird warnings or errors. Try some basic SkyEye commands such as a [RADIO CHECK](PLAYER.md#radio-check) and a [PICTURE](PLAYER.md#picture). Make sure the results you hear match what you see in the DCS F10 map.

## Reducing the Bill

You pay for the SkyEye server on an hourly basis. You can delete the server when you're not playing DCS to reduce your bill. Note that it's not enough to power off the server; you must delete it. 

You can recreate the server at any time by following the steps above; if you saved the customized `cloud-config.yaml` file, you can recreate the server in a few clicks. If you're an advanced user, see the [autoscaling documentation](ADMIN.md#autoscaling-experimental) for a way to automate this task.

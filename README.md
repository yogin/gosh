
# gosh

GoSH (Go-SHell, GO-sSH) is a terminal based utility to list instances in cloud providers and help connect to them through ssh.

## Supported cloud providers

* AWS (ec2)
* ...

More providers to be added in the future.

## Usage

`gosh` does its best to use configured tools in your environments.

To access the AWS API, it relies on the [AWS CLI config/credentials](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html).

To connect to an instance, it simply calls `ssh <ip>` and relies on your system's [ssh client configuration](https://www.ssh.com/academy/ssh/config). By default it will use the instance's private IP, but this behavior can be overridden to use the public IP instead through the configuration file.

Here's a simple example on how to configure the ssh client's config file to be able to connect to IPs in the `10.0.0.0/16` block. See the link above for more options.

```
Host 10.0.*
  User tester
  IdentityFile ~/.ssh/tester.id_rsa
```

When `gosh` starts without any configuration, it will try to use the `default` profile for the AWS CLI to fetch instances.

If you need more than just the default, you can hit `w` once `gosh` is running to write a configuration file to `~/.gosh.yaml` (user's home directory).

## Configuration

Here's an example of a configuration file that has 2 profiles `usw1` and `use1`. Both profiles use the `default` profile from AWS CLI but with different regions. The `refresh` tells `gosh` to pull the list of instances at regular intervals (in seconds)

```yaml
version: 1
profiles:
    - id: usw1
      provider: aws
      name: default
      region: us-west-1
      prefer_public_ip: false
      refresh:
        enabled: true
        interval: 60
    - id: use1
      provider: aws
      name: default
      region: us-east-1
      prefer_public_ip: false
      refresh:
        enabled: true
        interval: 60
show_utc_time: true
show_local_time: true
time_format: "2006-01-02 15:04:05"
```

When `gosh` starts it will look for a configuration files in this order:

1. `./.gosh.yaml`
2. `~/.gosh.yaml`
3. `/etc/gosh.yaml`

When saving the configuraiton it will overwrite the file that it loaded, or if none were found it will write it to `~/.gosh.yaml`.

## Keybinds

`gosh` has various keybinds to navigate the UI:

* `<tab>` to cycle through profile pages
* `1` through `9` for quick access to profile pages
* `q` to exit (ctrl-c works also)
* `w` to save the configuration file
* `r` to refresh instances in the current profile
* `R` to toggle automatic refreshes for the current profile
* `up/down/left/right` (`hjkl`) to navigate through individual instances and colums
* `pageUp/pageDown/home/end` to quick navigation through the list of instances
* `ENTER` to ssh into the current selected instance
* `~` to toggle display of an internal log (only needed for development)

## Upcoming

Here's a non-exhaustive list of things planned:

* Add configuration options for which instances fields and tags to display
* Disable the internal log and log window by default and use a configuration option
* Allow configuring the refresh interval through the UI
* Add search and filtering capabilities to quickly find instances
* ...

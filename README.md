# hidemego - Network Anonymization Tool for Linux

## What is hidemego?

hidemego is a network application useful to anonymize the traffic of linux servers and workstations. 

The action hidemego performs is to redirect all traffic via SOCKS5 (Tor) Proxy using the Transparent Proxy Method.

DNS requests are also anonymized and hidemego makes DNS Leak almost impossible.

## When to use hidemego?

hidemego can be used under any circumstances that require a mandatory anonymity requirement.

The use cases could be different.

  - Bypassing Censorship
  - Testing
  - Scraping
  - Preventing attacks on Servers with Critical Information
  - Communicating in total anonymity through the programs of daily use.

The creator's hope is that it will be used to improve people's privacy without violating any law.

## Compatibility

The compatibility of hidemego is verified on all RHEL based distributions such as Fedora, CentOS and Rocky. 

## Requirements

To use hidemego you need a Linux distribution with:

- Tor
- systemd
- NetworkManager
- iptables
- ip (command)
- ethtools
- awk
- netcat

## How Can I Install hidemego from source on Linux?

To compile hidemego from source you need:

- Go (1.17)

**Installation from terminal**

`$ git clone https://github.com/multiversecoder/hidemego hidemego && cd hidemego`

In the hidemego folder use the `go install` to install the software

`$ go build -o hidemego`

`$ sudo cp hidemego /usr/local/bin/hidemego`

To install man pages

`$ sudo cp ./hidemego.1 /usr/local/share/man/man1/hidemego.1`

## Uninstall hidemego & hidemego man pages

Run these commands to uninstall hidemego and man pages

`$ sudo rm /usr/local/bin/hidemego && sudo rm /usr/local/share/man/man1/hidemego.1`

## Features

- Single Executable
- Ease of use
- MAC address spoofing
- Compatibility with SELinux
- Security against DNS Leaks
- No need to use external libraries

## Usage

To start hidemego without special configurations use the command:
    
`$ sudo hidemego start`
    
To start hidemego in stealth mode to change the MAC Address of the interfaces, use the command:
    
`$ sudo hidemego start -ifaces="enp1s0"`
    
To end the hidemego anonymisation session, use the command:

`$ sudo hidemego stop`

NOTES:
    
    - <interface(s)> MUST BE ADDED as comma separed list or as string if you need to spoof the MAC address of only one interface


## Optional hidemego arguments:

Usage of hidemego:
  
  
  -id int
      Tor user id. If no value is passed. Hidemego will parse default-torrc to identify user and related id
  
  -ifaces string
      Interfaces that must change MAC Address (separed by comma)
  
  -no14
      Excludes Nodes from 14 eyes countries
  
  -no14p
      Excludes Nodes from 14 eyes countries plus other dangerous countries
  
  -no5
      Excludes Nodes from 5 eyes countries
  
  -no9
      Excludes Nodes from 9 eyes countries
  
  -pass string
      The Tor Control Authentication Password

  -cport int
      Tor Control Port for 127.0.0.1 (default 9052)
  
  -dport int
      DNS Port (default 5354)
  
  -saport int
      Socks Authentication Port for 127.0.0.1.0 (default 9151)
  
  -sdport int
      Socks Destination Port for 127.0.0.1.1 (default 9051)
  
  -tport int
      Tor Transport Port (default 9040)
  
  -user string
      Tor process user name. If no value is passed, Hidemego will parse defaults-torrc to identify user

  
## Finding your Tor ID

If the `-id` flag is not passed to the executable, the software will be able to automatically find the Tor user id automatically using /usr/share/tor/defaults-torrc

If you want to use a custom id, to identify the id to pass as `id` flag, from the terminal run:
    
`$ id -u (Tor username)`
    
Finding ID of Default Tor User on RHEL/CentOS/Fedora:

`$ id -u toranon`

Finding ID of Default Tor User on Debian/Ubuntu/Mint:

`$ id -u debian-tor`

Finding ID of Default Tor User on ARCH:

`$ id -u tor`

# DISCLAIMER
    
The author of this software assumes no responsibility for the use of this software to perform actions that do not comply with the law and/or damage property and/or individuals.
Using this software you take full responsibility for your actions.

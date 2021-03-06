# Systems- PING

## What is it?

Small Ping CLI application for MacOS or Linux. The CLI app accepts a hostname or an IP address as its argument along with options to provide TTL and Dealy values.The app emits "echo requests" in an infinite loop with peridoic delay. It report loss and RTT times for each sent message.

## Libraries
The application is Built in Go. Packages used are: icmp, ipv4, ipv6

## Build
download and install packages and dependencies
```
go get golang.org/x/net/icmp
go get golang.org/x/net/ipv4
go get golang.org/x/net/ipv6

```
Build command
```
go build ping.go

```
## Usage
The app need to be run using previlged user. Please use run the command with 'sudo'

```
sudo ./ping [-ttl TTLValue][-d delay_between_messages] <IP address/hostname>
```
The CLI app supports both IPV4 and IPV6 addresses. The TTL(Time to live) and delay values can be provided using -ttl , -d options respectively. if no options are provided, The app runs with default values of 64 for ttl and 1s dealy.   
Examples:
```
sudo ./ping 8.8.8.8
sudo ./ping -ttl=50 -d=1s 8.8.8.8
sudo ./ping -ttl=30 -d=2s www.cloudflare.com
sudo ./ping -ttl=2 -d=2s ipv6.google.com
sudo ./ping -d=300ms www.google.com
sudo ./ping -ttl=100 -d=5s www.facebook.com
```
## Extra Credit

1. Added support for both IPv4 and IPv6.
2. App allows to set TTL as an argument and reports the corresponding "time exceeded” ICMP messages.
3. Added -d option to allow the delay between messages as an argument.

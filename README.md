# DoDoH-PoC
Local resolver using DNS-over-HTTPS Proof of Concept

> This is a Proof-of-Concept code! DO NOT USE IN PRODUCTION

## What and why

This is a simple DNS proxy server that runs locally, and forwards any DNS requests system-wide to specified DNS-over-HTTPS server.

Made to refresh low-level DNS knowledge, and understand how DoH works in practice.

## How to build and run

On linux environment:

```
git clone https://github.com/mdukat/DoDoH-PoC.git
cd DoDoH-PoC
go mod tidy
go build .
sudo ./DoDoH-PoC
```

Or using Docker:

```
git clone https://github.com/mdukat/DoDoH-PoC.git
cd DoDoH-PoC
docker build -t dodoh-dns .
docker run -p 127.4.4.4:53:53/udp dodoh-dns
```

### NetworkManager configuration

Simply change your DNS server configuration to `127.4.4.4`.

### `/etc/resolv.conf`

```
nameserver 127.4.4.4
```

## Future plans

Build production-ready version, with TLS certificate pinning, cache, and user-friendly configuration.

## Problems found during research

 - Go does not provide RFC-1035 compliant DNS message structures and methods, except for their "outside main Go tree" repository: https://pkg.go.dev/golang.org/x/net@v0.46.0/dns/dnsmessage
 - There's no simple way of pinning TLS certificates, as [tam7t/hpkp](https://github.com/tam7t/hpkp) hasn't been updated in 9 years, yet it's a first hit on pkg.go.dev
   - It also has problems with forcing HTTP/1, which I've been fighting with and failing to fix in a timeframe I gave myself for this project
 - [miekg/dns](https://codeberg.org/miekg/dns) has split into v1 (on github) and v2 (on codeberg). v2 feels somewhat more complicated to get into for someone (like me), who never worked with DNS in Go.

## Some data

### Standards are great

Thanks to DoH conforming to RFC-1035 DNS message structure, we are able to pass any request to next "hop" without modification, and get proper response, no matter query type:

```
dell5290:~/dns-test/DoDoH-PoC$ dig onet.pl txt

; <<>> DiG 9.20.13 <<>> onet.pl txt
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 40692
;; flags: qr rd ra; QUERY: 1, ANSWER: 6, AUTHORITY: 0, ADDITIONAL: 1

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 1232
; COOKIE: 764d28e18e4e0ff10100000068f4ba062409a4a864f879be (good)
;; QUESTION SECTION:
;onet.pl.			IN	TXT

;; ANSWER SECTION:
onet.pl.		300	IN	TXT	"facebook-domain-verification=5g4w0j45neyxsoxyqp23mhp2a7i9hu"
onet.pl.		300	IN	TXT	"tiktok-developers-site-verification=vBe6AC2lJx5khgomOpLopWZj6wM7GR0h"
onet.pl.		300	IN	TXT	"v=spf1 ip4:213.180.128.0/19 ip4:141.105.16.0/27 include:amazonses.com -all"
onet.pl.		300	IN	TXT	"MS=ms47770053"
onet.pl.		300	IN	TXT	"mojecertpl-site-verification-9fKJjs7dgM9XsuYtcd5DeU10cNDdikMX"
onet.pl.		300	IN	TXT	"google-site-verification=r9Ke_qyyz-v1Dcz92qf7IGqlU3mTe-muUWnPi9Kmnyw"

;; Query time: 149 msec
;; SERVER: 127.4.4.4#53(127.4.4.4) (UDP)
;; WHEN: Sun Oct 19 12:14:30 CEST 2025
;; MSG SIZE  rcvd: 527
```

### Security

By passing our system-wide DNS via specific DoH-server, we are able to secure our environment from DNS Leaks. Better yet, we are able to block all outgoing UDP/53 requests with firewall, to gain control over apps which might use public DNS servers directly.

Tested with:

 - https://www.dnsleaktest.com/
 - https://mullvad.net/en/check

### Speed

Of course it's waaaaaaay slower than standard, unencrypted DNS. I won't get into specifics of how slow it is exactly, but when browsing internet and doing "work stuff", I don't see any difference really.

### And as always

[The more you fck around, the more you find out.](https://www.youtube.com/watch?v=Q0otmkns_Aw)


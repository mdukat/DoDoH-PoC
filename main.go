/*
MIT License

Copyright (c) 2025 Mateusz Dukat

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
	"bytes"
	"log"
	"fmt"
	"io"
	"net/http"
	"crypto/tls"
	"os"
	"time"

	// If you want to build proper proxy, I would suggest to use a newer release of miekg/dns library linked here: https://codeberg.org/miekg/dns
	"github.com/miekg/dns"
)

func askMullvad(m dns.Msg) *dns.Msg {
	// In this PoC we use Mullvad DoH service:
	// https://mullvad.net/en/help/dns-over-https-and-dns-over-tls

	// and we use IPv4, so we dont have to ask another resolver for 'dns.mullvad.net'
	//dohURL := "https://dns.mullvad.net/dns-query"
	dohURL := "https://194.242.2.2/dns-query"

	packed, err := m.Pack()
	if err != nil {
		fmt.Fprintln(os.Stderr, "pack error:", err)
		os.Exit(1)
	}

	// HTTP POST with application/dns-message (RFC 8484)
	req, err := http.NewRequest(http.MethodPost, dohURL, bytes.NewReader(packed))
	if err != nil {
		fmt.Fprintln(os.Stderr, "request error:", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/dns-message")
	req.Header.Set("Accept", "application/dns-message")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			// We need to disable TLS verification, because we ask directly to IPv4 address which is not in TLS certificate provided by dns.mullvad.net :(
			// TODO: Certificate pinning
			InsecureSkipVerify: true,
		},
		// For some reason, Go 1.25.1 wants to use HTTP/1 by default on my machine
		ForceAttemptHTTP2: true,
	}
	client := &http.Client{
		Timeout: 7 * time.Second,
		Transport: tr,
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "client.Do error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "unexpected status %d: %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "https response read error:", err)
		os.Exit(1)
	}

	// parse DNS response
	rm := new(dns.Msg)
	if err := rm.Unpack(body); err != nil {
		fmt.Fprintln(os.Stderr, "unpack error:", err)
		os.Exit(1)
	}
	return rm
}

func main() {
	// Catch all requests and handle them via askMullvad()
	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		if len(r.Question) == 0 {
			// Return format error if there is no question in request
			m := new(dns.Msg)
			m.SetRcode(r, dns.RcodeFormatError)
			_ = w.WriteMsg(m)
			return
		}

		// Log request for domain name
		q := r.Question[0]
		domain := q.Name
		log.Printf("Query for: %s (type %d)\n", domain, q.Qtype)

		// Forward question to mullvad and wait for response
		m := askMullvad(*r)

		// Send response back to client
		if err := w.WriteMsg(m); err != nil {
			log.Printf("WriteMsg error: %v", err)
		}
	})

	// UDP server on :53
	// If you want to run it with lower privileges than root, you'll have to
	// do some forwarding magic using iptables between udp:53 and whatever
	// port you'll want to run this server on

	// I guess, modern solution would be to use docker, as it's daemon runs as root
	// https://stackoverflow.com/questions/27596409/how-do-i-publish-a-udp-port-on-docker
	server := &dns.Server{Addr: "127.4.4.4:53", Net: "udp"}
	log.Printf("Starting DNS server on %s\n", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}


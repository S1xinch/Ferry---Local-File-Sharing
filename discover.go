package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Finding other computers on the same network. A computer sharing a folder
// answers a short "who is sharing?" message on a fixed port. The asking side
// sends the message to every address on the local network, so no special
// network permissions (like broadcast) are needed. Replies never include the PIN.

const (
	discoverPort = 47616
	discoverMsg  = "FERRY-DISCOVER"
)

// discovered is one computer that answered.
type discovered struct {
	Name   string `json:"name"`   // this computer's name
	Folder string `json:"folder"` // name of the folder it shares
	Addr   string `json:"addr"`   // host:port to type into Receive
}

type discoverReply struct {
	Name   string `json:"name"`
	Folder string `json:"folder"`
	Port   int    `json:"port"`
}

// answerDiscovery makes this computer findable while it shares a folder.
// It returns a function that stops answering.
func answerDiscovery(folder string, port int) func() {
	pc, err := net.ListenPacket("udp4", fmt.Sprintf(":%d", discoverPort))
	if err != nil {
		// Another Ferry is already answering on this computer, or the port is busy.
		return func() {}
	}
	host, _ := os.Hostname()
	reply, _ := json.Marshal(discoverReply{Name: host, Folder: filepath.Base(folder), Port: port})
	go func() {
		buf := make([]byte, 256)
		for {
			n, from, err := pc.ReadFrom(buf)
			if err != nil {
				return
			}
			if string(buf[:n]) == discoverMsg {
				pc.WriteTo(reply, from)
			}
		}
	}()
	return func() { pc.Close() }
}

// localTargets lists the addresses to ask: every address in the /24 around
// each local IPv4 address, plus this computer itself.
func localTargets() []*net.UDPAddr {
	targets := []*net.UDPAddr{{IP: net.IPv4(127, 0, 0, 1), Port: discoverPort}}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return targets
	}
	seen := map[string]bool{}
	for _, a := range addrs {
		ipn, ok := a.(*net.IPNet)
		if !ok || ipn.IP.To4() == nil || ipn.IP.IsLoopback() {
			continue
		}
		base := ipn.IP.To4().Mask(net.CIDRMask(24, 32))
		start := binary.BigEndian.Uint32(base)
		for i := uint32(1); i < 255; i++ {
			ip := make(net.IP, 4)
			binary.BigEndian.PutUint32(ip, start+i)
			key := ip.String()
			if seen[key] {
				continue
			}
			seen[key] = true
			targets = append(targets, &net.UDPAddr{IP: ip, Port: discoverPort})
		}
	}
	return targets
}

// findComputers asks the local network who is sharing, and returns the
// computers that answered within the wait time.
func findComputers(wait time.Duration) ([]discovered, error) {
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	for _, t := range localTargets() {
		conn.WriteTo([]byte(discoverMsg), t)
	}

	conn.SetReadDeadline(time.Now().Add(wait))
	var out []discovered
	seen := map[string]bool{}
	buf := make([]byte, 512)
	for {
		n, from, err := conn.ReadFrom(buf)
		if err != nil {
			break // the wait time is over
		}
		var r discoverReply
		if json.Unmarshal(buf[:n], &r) != nil || r.Port == 0 {
			continue
		}
		host, _, err := net.SplitHostPort(from.String())
		if err != nil {
			continue
		}
		addr := net.JoinHostPort(host, strconv.Itoa(r.Port))
		if seen[addr] {
			continue
		}
		seen[addr] = true
		out = append(out, discovered{Name: r.Name, Folder: r.Folder, Addr: addr})
	}
	return out, nil
}

func (a *app) discover(r *http.Request) (any, error) {
	found, err := findComputers(2 * time.Second)
	if err != nil {
		return nil, err
	}
	return map[string]any{"computers": found}, nil
}

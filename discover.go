package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
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
// sends the message to every address on the local network and to its
// broadcast address. Replies never include the PIN.

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

// localTargets lists the addresses to ask on each network this computer is
// on: every address in the network (capped to the 1022 around this computer
// on very large networks) plus the network's broadcast address, and finally
// this computer itself. VPN links are skipped.
func localTargets() (lan []*net.UDPAddr, self *net.UDPAddr) {
	self = &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: discoverPort}
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, self
	}
	seen := map[string]bool{}
	add := func(ip net.IP) {
		if key := ip.String(); !seen[key] {
			seen[key] = true
			lan = append(lan, &net.UDPAddr{IP: ip, Port: discoverPort})
		}
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&(net.FlagLoopback|net.FlagPointToPoint) != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, a := range addrs {
			ipn, ok := a.(*net.IPNet)
			if !ok || ipn.IP.To4() == nil || !ipn.IP.IsPrivate() {
				continue
			}
			ones, _ := ipn.Mask.Size()
			if ones < 22 {
				ones = 22
			}
			if ones > 30 {
				continue
			}
			base := binary.BigEndian.Uint32(ipn.IP.To4().Mask(net.CIDRMask(ones, 32)))
			size := uint32(1) << (32 - ones)
			for i := uint32(1); i < size-1; i++ {
				ip := make(net.IP, 4)
				binary.BigEndian.PutUint32(ip, base+i)
				add(ip)
			}
			// The network's own broadcast address, for routers that allow it.
			full := ipn.IP.To4().Mask(ipn.Mask)
			bcast := make(net.IP, 4)
			for i := range bcast {
				bcast[i] = full[i] | ^ipn.Mask[i]
			}
			add(bcast)
		}
	}
	if len(lan) > 0 {
		add(net.IPv4bcast)
	}
	return lan, self
}

// errNoLocalNetwork means every message to the local network was refused by
// this computer, usually because Ferry was not allowed to use it.
var errNoLocalNetwork = errors.New("Ferry could not reach your local network. " +
	"On a Mac, open System Settings > Privacy & Security > Local Network and turn Ferry on, then try again.")

// findComputers asks the local network who is sharing, and returns the
// computers that answered within the wait time. It asks twice, because the
// first message to a computer is sometimes lost while the network looks it up.
func findComputers(wait time.Duration) ([]discovered, error) {
	lc := net.ListenConfig{Control: allowBroadcast}
	conn, err := lc.ListenPacket(context.Background(), "udp4", ":0")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	lan, self := localTargets()
	send := func() (sent int) {
		conn.WriteTo([]byte(discoverMsg), self)
		for _, t := range lan {
			if _, err := conn.WriteTo([]byte(discoverMsg), t); err == nil {
				sent++
			}
		}
		return sent
	}
	if len(lan) == 0 {
		send()
	} else if send() == 0 {
		return nil, errNoLocalNetwork
	}
	go func() {
		time.Sleep(wait / 3)
		send()
	}()

	conn.SetReadDeadline(time.Now().Add(wait))
	var out []discovered
	seen := map[string]bool{}
	buf := make([]byte, 512)
	for {
		n, from, err := conn.ReadFrom(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				break // the wait time is over
			}
			if errors.Is(err, net.ErrClosed) {
				break
			}
			continue // e.g. "port unreachable" from a computer without Ferry
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

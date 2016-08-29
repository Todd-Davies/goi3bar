package network

import (
	i3 "github.com/ToddDavies/goi3bar"

	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"
)

const Identifier = "network"

const (
	notConnected = "Not connected"
	ethFormat    = "Connected: %v (%vMb/s)"
	ipAddrRxStr  = "inet\\s+((\\d{1,3}\\.){3}\\d{1,3})"
	isUpRxStr    = "state\\s+UP"
)

var (
	ipAddrRx = regexp.MustCompile(ipAddrRxStr)
	isUpRx   = regexp.MustCompile(isUpRxStr)
)

type NetworkDevice interface {
	i3.Generator

	FriendlyName() string
	Interface() string
	IP() net.IP
	Speed() uint64
	Connected() bool

	Update() error
}

// A BaseBasicNetworkDevice describes a network device to be displayed on an i3bar
// TODO: Usage speed?
type BasicNetworkDevice struct {
	// A friendly name to be used on the bar
	Name string

	// Name of the network interface for the corresponding device
	Identifier string

	ip net.IP
	// Link speed in kbits/sec
	speed     uint64
	connected bool
}

func (d *BasicNetworkDevice) FriendlyName() string {
	return d.Name
}

func (d *BasicNetworkDevice) Interface() string {
	return d.Identifier
}

func (d *BasicNetworkDevice) IP() net.IP {
	return d.ip
}

func (d *BasicNetworkDevice) Speed() uint64 {
	return d.speed
}

func (d *BasicNetworkDevice) Connected() bool {
	return d.connected
}

// Update implements NetworkDevice
func (d *BasicNetworkDevice) Update() error {
	out, err := exec.Command("ip", "addr", "show", d.Identifier).Output()
	if err != nil {
		return err
	}

	output := string(out)

	d.connected = isUpRx.MatchString(output)
	if !d.connected {
		return nil
	}

	matches := ipAddrRx.FindStringSubmatch(output)
	if len(matches) < 2 {
		d.ip = nil
	} else {
		d.ip = net.ParseIP(matches[1])
	}

	// TODO: Bring crushing reality upon our users of their network speed
	d.speed = 1000000

	return nil
}

// Generate implements i3.Generator
func (d *BasicNetworkDevice) Generate() ([]i3.Output, error) {
	d.Update()

	if !d.connected {
		return []i3.Output{{
			FullText: fmt.Sprintf(notConnectedTpl, d.Name),
			Color:    i3.DefaultColors.Crit,
		}}, nil
	}

	speed := strconv.FormatUint(d.speed/1000, 10)

	var ip string
	if d.ip != nil {
		ip = d.ip.String()
	} else {
		ip = "Acquiring IP"
	}

	text := fmt.Sprintf(ethFormat, ip, speed)

	return []i3.Output{{
		Name:      Identifier,
		Instance:  d.Identifier,
		FullText:  text,
		Color:     i3.DefaultColors.OK,
		Separator: true,
	}}, nil
}

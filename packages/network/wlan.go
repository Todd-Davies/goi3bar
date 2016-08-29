package network

import (
	i3 "github.com/ToddDavies/goi3bar"

	"fmt"
	"os/exec"
	"regexp"
	"strconv"
)

const (
	essidRxStr    = "ESSID:\"(.*)\""
	strengthRxStr = "Link Quality=(\\d+)/(\\d+)"
)

const (
	defaultTpl      = "%v: %v %v%% (%v)"
	noStrengthTpl   = "%v: %v (%v)"
	notConnectedTpl = "%v not connected"
)

var (
	essidRx    = regexp.MustCompile(essidRxStr)
	strengthRx = regexp.MustCompile(strengthRxStr)
)

type WLANDevice struct {
	BasicNetworkDevice

	WarnThreshold int
	CritThreshold int

	// Signal strength as a percentage.
	strength int
	// true if there is no signal strength available from iw
	strengthUnavailable bool
	essid               string
}

func (d *WLANDevice) updateESSID(input string) error {
	matches := essidRx.FindStringSubmatch(input)
	if len(matches) < 2 {
		return fmt.Errorf("Couldn't match ESSID")
	}

	d.essid = matches[1]

	return nil
}

func (d *WLANDevice) updateStrength(input string) error {
	matches := strengthRx.FindStringSubmatch(input)
	d.strengthUnavailable = len(matches) < 3
	if d.strengthUnavailable {
		return nil
	}

	strengthNum, errN := strconv.Atoi(matches[1])
	strengthDenom, errD := strconv.Atoi(matches[1])
	if errN != nil {
		return errN
	}
	if errD != nil {
		return errD
	}

	d.strength = (strengthNum * 100) / strengthDenom

	return nil
}

func (d *WLANDevice) fetch() (string, error) {
	output, err := exec.Command("iwconfig", d.Identifier).Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func (d *WLANDevice) update() error {
	d.BasicNetworkDevice.Update()

	iwOut, err := d.fetch()
	if err != nil {
		return err
	}

	err = d.updateStrength(iwOut)
	if err != nil {
		return err
	}

	err = d.updateESSID(iwOut)
	if err != nil {
		return err
	}

	return nil
}

// Generate implements Generator
func (d *WLANDevice) Generate() ([]i3.Output, error) {
	err := d.update()
	if err != nil {
		return nil, err
	}

	if !d.connected {
		return []i3.Output{{
			FullText: fmt.Sprintf(notConnectedTpl, d.Name),
			Color:    i3.DefaultColors.Crit,
		}}, nil
	}

	var ip string
	if d.ip == nil {
		ip = "Acquiring IP"
	} else {
		ip = d.ip.String()
	}

	var txt string
	if d.strengthUnavailable {
		txt = fmt.Sprintf(noStrengthTpl, d.Name, d.essid, ip)
	} else {
		txt = fmt.Sprintf(defaultTpl, d.Name, d.essid, d.strength, ip)
	}

	var color string
	switch {
	case d.strengthUnavailable:
		color = i3.DefaultColors.OK
	case d.strength < d.CritThreshold:
		color = i3.DefaultColors.Crit
	case d.strength < d.WarnThreshold:
		color = i3.DefaultColors.Warn
	default:
		color = i3.DefaultColors.OK
	}

	return []i3.Output{{
		Name:      Identifier,
		Instance:  d.Identifier,
		FullText:  txt,
		Color:     color,
		Separator: true,
	}}, nil
}

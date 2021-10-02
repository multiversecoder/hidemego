package tor

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/multiversecoder/hidemego/tools"
)

var (
	DefaultTorRC  = path.Join("/", "usr", "share", "tor", "defaults-torrc")
	HidemegoLib   = path.Join("/", "var", "lib", "tor", "hidemego")
	HidemegoTorRC = path.Join("/", "etc", "tor", "hidemego.torrc")
	ServiceName   = "tor@hidemego.service"
)

// Extract toruser from defaults-torrc
func Usr() (string, error) {
	f, err := ioutil.ReadFile(DefaultTorRC)
	if err != nil {
		return "", err
	}
	for _, data := range strings.Split(string(f), "\n") {
		if strings.Contains(data, "User") {
			return strings.Split(data, " ")[1], nil
		}
	}
	return "", fmt.Errorf("cannot find the right tor username")
}

func Restart(reload ...bool) error {
	var (
		relo   bool
		action string
	)
	if len(reload) > 0 {
		relo = reload[0]
	}
	switch relo {
	case true:
		action = "reload"
	case false:
		action = "restart"
	}
	return exec.Command("systemctl", action, ServiceName).Run()
}

func Stop() error {
	return exec.Command("systemctl", "stop", ServiceName).Run()
}

func ChangeControlIdentity(ip string, tpass string, tport int) error {
	m := make(map[string]interface{})
	m["TorPassword"] = tpass
	m["TorPort"] = tport
	tb, err := tools.Read("nymn", m)
	if err != nil {
		return err
	}
	script, err := tools.TempFile("./nymn.sh", tb.Bytes())
	if err != nil {
		return err
	}
	defer os.Remove(script)
	netcat := exec.Command("/bin/bash", script)
	if err := netcat.Run(); err != nil {
		return err
	}
	return nil
}

// Change IP Addr sending HUP to tor
func ChangeHUPIdentity(ip string) error {
	pid, err := exec.Command("pidof", "tor").Output()
	if err != nil {
		return err
	}
	hup := exec.Command("kill", "-HUP", strings.TrimSuffix(string(pid), "\n"))
	if err := hup.Run(); err != nil {
		return err
	}
	return nil
}

func ChangeIdentity(tpass string, tport int) (string, error) {
	var hip string
	ip, err := tools.GetIPAddress()
	if err != nil {
		return "", err
	}
	err = ChangeHUPIdentity(ip)
	if err != nil {
		return "", err
	}
	hip, err = tools.GetIPAddress()
	if err != nil {
		return "", err
	}
	if ip == hip && tpass != "" {
		time.Sleep(3 * time.Second)
		err := ChangeControlIdentity(ip, tpass, tport)
		if err != nil {
			return "", err
		}
		hip, err = tools.GetIPAddress()
		if err != nil {
			return "", err
		}
	}
	return hip, nil
}

func ID() (int, error) {
	var torid int
	torusr, err := Usr()
	if err != nil {
		return torid, err
	}
	id, err := exec.Command("id", "-ur", torusr).Output()
	if err != nil {
		return torid, err
	}
	torid, _ = strconv.Atoi(strings.ReplaceAll(string(id), "\n", ""))
	return torid, nil
}

func ChangeDirOwner(dir string) error {
	torusr, err := Usr()
	if err != nil {
		return err
	}
	owner := fmt.Sprintf("%s:%s", torusr, torusr)
	err = exec.Command("chown", "-R", owner, dir).Run()
	if err != nil {
		return err
	}
	return nil

}

func SetTorRC(countries string, torPort, socksDestPort, socksAuthPort, controlPort, dnsPort int, tuser string, tpass ...string) error {
	var tb bytes.Buffer
	var m = make(map[string]interface{})
	m["TorPort"] = torPort
	m["Countries"] = countries
	m["DataDir"] = HidemegoLib
	if len(tpass) > 0 {
		m["ControlPort"] = controlPort
		m["HasTorControl"] = true
		m["TPass"] = tpass[0]
	} else {
		m["ControlPort"] = controlPort
		m["HasTorControl"] = false
		m["TPass"] = ""
	}
	m["SocksDestPort"] = socksDestPort
	m["SocksAuthPort"] = socksAuthPort
	m["DNSPort"] = dnsPort
	tb, err := tools.Read("torrc", m)
	if err != nil {
		return err
	}
	os.Mkdir(HidemegoLib, 0777)
	chown := exec.Command("chown", tuser+":root", HidemegoLib)
	if err := chown.Run(); err != nil {
		return err
	}
	return ioutil.WriteFile(HidemegoTorRC, tb.Bytes(), 0644)
}

func RemoveTorRc() error {
	if _, err := os.Stat(HidemegoTorRC); !os.IsNotExist(err) {
		return os.Remove(HidemegoTorRC)
	}
	return nil
}

func RemoveHideMeGoDir() error {
	if _, err := os.Stat(HidemegoLib); !os.IsNotExist(err) {
		return os.RemoveAll(HidemegoLib)
	}
	return nil
}


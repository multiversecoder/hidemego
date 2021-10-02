package linux

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/multiversecoder/hidemego/tools"
)

var (
	previousSysctlConf = path.Join(os.Getenv("HOME"), ".config", "hidemego", "prev.sysctl.conf")
	ipCommand, _       = tools.Which("ip")
	ipTablesCommand, _ = tools.Which("iptables")
	resolvConf         = path.Join("/", "etc", "resolv.conf")
)

func DefaultMacAddr(iface string) (string, error) {
	cmd, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("ethtool -P %s | awk '{print $3}'", iface)).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(strings.TrimSuffix(string(cmd), "\n")), nil
}

func RestartNetwork(reload ...bool) error {
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
	return exec.Command("systemctl", action, "NetworkManager.service").Run()
}

func DefaultIfaces() ([]string, error) {
	tb, err := tools.Read("getifaces", map[string]interface{}{"Ip": ipCommand})
	if err != nil {
		return nil, err
	}
	cmd, err := exec.Command("/bin/sh", "-c", tb.String()).Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSuffix(string(cmd), "\n"), "\n"), nil
}

func HasIface(iface string) bool {
	ifaces, err := DefaultIfaces()
	if err != nil {
		return false
	}
	for _, i := range ifaces {
		if iface == i {
			return true
		}
	}
	return false
}

func SELManage(port int, add bool, dns ...bool) error {
	var (
		action string
		err    error
		isDNS  bool
	)
	if len(dns) > 0 {
		isDNS = dns[0]
	}
	switch add {
	case false:
		action = "-d"
	case true:
		action = "-a"
	}

	if isDNS {
		err = exec.Command("semanage", "port", action, "-t", "dns_port_t", "-p", "tcp", strconv.FormatInt(int64(port), 10)).Run()
		if err != nil {
			return err
		}
		err = exec.Command("semanage", "port", action, "-t", "dns_port_t", "-p", "udp", strconv.FormatInt(int64(port), 10)).Run()
		if err != nil {
			return err
		}
	} else {
		err = exec.Command("semanage", "port", action, "-t", "tor_port_t", "-p", "tcp", strconv.FormatInt(int64(port), 10)).Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func HasSELPort(needle int) bool {
	cmd, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("semanage port -l | grep %d", needle)).Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.TrimSuffix(string(cmd), "\n"), strconv.FormatInt(int64(needle), 10))
}

func SetIPTablesRules(nontor string, torid, port, dnsPort int) error {
	var tb bytes.Buffer
	var m = make(map[string]interface{})
	m["IPTables"] = ipTablesCommand
	m["ExcludedTorAddrs"] = nontor
	m["TorID"] = torid
	m["TorPort"] = port
	m["DNSPort"] = dnsPort
	m["IfaceIF"] = "wlo1"
	m["IfaceOF"] = "wlo1"
	tb, err := tools.Read("iptr", m)
	if err != nil {
		return err
	}
	script, err := tools.TempFile("hidemego_iptables", tb.Bytes())
	if err != nil {
		return err
	}
	defer os.Remove(script)
	cmd := exec.Command("chmod", "+x", script)
	if err := cmd.Run(); err != nil {
		return err
	}
	ipr := exec.Command("/bin/bash", script)
	if err := ipr.Run(); err != nil {
		return err
	}
	return nil
}

func FlushIPTablesRules() error {
	var tb bytes.Buffer
	var m = make(map[string]interface{})
	m["IPTables"] = ipTablesCommand
	tb, err := tools.Read("iptf", m)
	if err != nil {
		return err
	}
	script, err := tools.TempFile("hidemegp_flush_iptables", tb.Bytes())
	if err != nil {
		return err
	}
	defer os.Remove(script)
	cmd := exec.Command("chmod", "+x", script)
	if err := cmd.Run(); err != nil {
		return err
	}
	flush := exec.Command("/bin/bash", script)
	if err := flush.Run(); err != nil {
		return err
	}
	return nil
}

func IPSet(iface string, mode string) error {
	return exec.Command(ipCommand, "link", "set", iface, mode).Run()
}

func IPSetMACAddr(iface, mac string) error {
	return exec.Command(ipCommand, "link", "set", iface, "address", mac).Run()
}

func SetResolvConf() error {
	of, err := ioutil.ReadFile(resolvConf)
	if err != nil {
		return err
	}
	f := tools.IPRgx.ReplaceAll(of, []byte("127.0.0.1"))
	if !tools.SameFile(f, of) {
		if err := ioutil.WriteFile(resolvConf+".orig", of, 0644); err != nil {
			return err
		}
	}
	if err := ioutil.WriteFile(resolvConf,
		tools.IPRgx.ReplaceAll(f, []byte("127.0.0.1")),
		0644); err != nil {
		return err
	}
	return nil
}

func RestoreResolvConf() error {
	orig := resolvConf + ".orig"
	if _, err := os.Stat(orig); !os.IsNotExist(err) {
		f, err := ioutil.ReadFile(orig)
		if err != nil {
			return err
		}
		if _, err := os.Stat(resolvConf); !os.IsNotExist(err) {
			if err := os.Remove(resolvConf); err != nil {
				return err
			}
		}
		if err := os.Remove(orig); err != nil {
			return err
		}
		return ioutil.WriteFile(resolvConf, f, 0644)

	}
	return fmt.Errorf("%s.orig does not exists", resolvConf)
}

func PrepareLinuxKernel() {
	// diable ipv6
	tools.SetSysctl("net.ipv6.conf.lo.disable_ipv6=1")
	tools.SetSysctl("net.ipv6.conf.all.disable_ipv6=1")
	tools.SetSysctl("net.ipv6.conf.default.disable_ipv6=1")
	// disable kernel ip forwarding
	tools.SetSysctl("net.ipv4.ip_forward=0")
	// ignome icmp echo packets
	tools.SetSysctl("net.ipv4.icmp_echo_ignore_all=1")
	//tcp_mut_probing
	tools.SetSysctl("net.ipv4.tcp_mtu_probing=1")
	// prevent timestamp packet leakage
	tools.SetSysctl("net.ipv4.tcp_timestamps=0")
	// prevent assasination
	tools.SetSysctl("net.ipv4.tcp_rfc1337=1")
}

func SaveKernelConfigs() error {

	if _, err := os.Stat(previousSysctlConf); os.IsNotExist(err) {
		m := map[string]interface{}{
			"NetLoDisableIPv6":      tools.Sysctl("net.ipv6.conf.lo.disable_ipv6"),
			"NetAllDisableIPv6":     tools.Sysctl("net.ipv6.conf.all.disable_ipv6"),
			"NetDefaultDisableIPv6": tools.Sysctl("net.ipv6.conf.default.disable_ipv6"),
			"NetForwardIPv6":        tools.Sysctl("net.ipv4.ip_forward"),
			"NetICMPIgnoreAll":      tools.Sysctl("net.ipv4.icmp_echo_ignore_all"),
			"NetTCPMtuProbing":      tools.Sysctl("net.ipv4.tcp_mtu_probing"),
			"NetDisableTCPTS":       tools.Sysctl("net.ipv4.tcp_timestamps"),
			"NetRFC1337":            tools.Sysctl("net.ipv4.tcp_rfc1337")}
		tb, err := tools.Read("sysctl", m)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(previousSysctlConf, tb.Bytes(), 0644); err != nil {
			return err
		}
	}
	return nil
}

func RestoreKernelConfig() error {
	if _, err := os.Stat(previousSysctlConf); !os.IsNotExist(err) {
		b, err := ioutil.ReadFile(previousSysctlConf)
		if err != nil {
			return err
		}
		s := strings.Split(string(b), "\n")
		for _, r := range s {
			tools.SetSysctl(r)
		}
		return nil
	}
	return nil
}

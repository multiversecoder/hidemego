package tools

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"embed"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"text/template"
	"time"
)

var (
	IPRgx              = regexp.MustCompile(`[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}`)
	aRgx               = regexp.MustCompile(`("[^"]*"|[^"\s]+)(\s+|$)`)
	TorProjectCheckURL = "https://check.torproject.org"
	flagsFile          = path.Join(os.Getenv("HOME"), ".config", "hidemego", "hidemego.flags")
	//go:embed resources
	resources embed.FS
	templates = map[string]string{
		"torrc":     "resources/torrc.tmpl",
		"nymn":      "resources/nymn.tmpl",
		"iptr":      "resources/iptr.tmpl",
		"iptf":      "resources/iptf.tmpl",
		"getifaces": "resources/getifaces.tmpl",
		"getos":     "resources/getos.sh",
		"sysctl":    "resources/sysctl.tmpl"}

	// client for tor requests
	client = &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 60 * time.Second,
			}).Dial,
			// wait more because of tor
			TLSHandshakeTimeout: 60 * time.Second,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				CipherSuites: []uint16{
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				}}}}
)

func Read(t string, m map[string]interface{}) (bytes.Buffer, error) {
	var tb bytes.Buffer
	tpl, err := template.ParseFS(resources, templates[t])
	if err != nil {
		return tb, err
	}
	if err := tpl.Execute(&tb, m); err != nil {
		return tb, err
	}
	return tb, nil
}

func TempFile(name string, content []byte) (string, error) {
	f, err := ioutil.TempFile(os.TempDir(), name)
	if err != nil {
		return "", err
	}
	if _, err := f.Write(content); err != nil {
		return "", err
	}
	return f.Name(), nil
}

// Wait for internet connection
func CheckConn() bool {
	for {
		// connect to check.toprproject.com
		req, _ := http.NewRequest("GET", TorProjectCheckURL, nil)
		_, err := client.Do(req)
		if err != nil {
			// error missing internet connection
			continue
		}
		return true
	}
}

// Get Public IP Address from https://ipinfo.io/ip
func GetIPAddress() (string, error) {
	req, _ := http.NewRequest("GET", TorProjectCheckURL, nil)
	rsp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return "",
			fmt.Errorf(
				"fatal error while checking ip: %s", http.StatusText(rsp.StatusCode))
	}
	b, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}
	return string(IPRgx.Find(b)), nil
}

func SameFile(o, b []byte) bool {
	return bytes.Equal(o, b)
}

func RandMACAddr() (string, error) {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	buf[0] = (buf[0] | 2) & 0xfe
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5]), nil
}

func IsRoot() bool {
	return os.Getuid() == 0
}

func SetSysctl(s string) error {
	return exec.Command("sysctl", s).Run()
}

func Sysctl(command string) string {
	cmd, err := exec.Command("sysctl", command).Output()
	if err != nil {
		panic(err)
	}
	return strings.TrimSuffix(string(cmd), "\n")
}

func Which(s string) (string, error) {
	cmd, err := exec.Command("which", s).Output()
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(string(cmd), "\n", ""), nil
}

func Exists(s string) (bool, error) {
	w, err := Which(s)
	if err != nil {
		return false, err
	}
	return w != "", nil
}

func SaveFlags(args []string) error {
	s := strings.Join(args, " ")
	return ioutil.WriteFile(flagsFile, []byte(s), 0644)
}

func PreviousArgs() ([]string, error) {
	if _, err := os.Stat(flagsFile); os.IsNotExist(err) {
		return []string{}, nil
	}
	b, err := ioutil.ReadFile(flagsFile)
	if err != nil {
		return nil, err
	}
	return aRgx.FindAllString(string(b), -1), nil

}

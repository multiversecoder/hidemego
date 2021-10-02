package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/multiversecoder/hidemego/linux"
	"github.com/multiversecoder/hidemego/tools"
	"github.com/multiversecoder/hidemego/tor"
)

var (
	logger                = log.New(os.Stdout, "hidemego ", log.Ldate|log.Ltime)
	fl                    = flag.NewFlagSet("hidemego", flag.ExitOnError)
	torPass               string
	torUser               string
	torID                 int
	torPort               int
	socksDestPort         int
	socksAuthPort         int
	controlPort           int
	dnsPort               int
	no5, no9, no14, no14p bool
	ifaces                string
	once                  = sync.Once{}
	nokch                 bool
	confDir               = path.Join(os.Getenv("HOME"), ".config", "hidemego")
)

func init() {

}

func initialize() {
	if _, err := os.Stat(confDir); os.IsNotExist(err) {
		logger.Println("Creating Hidemego Config Directory")
		os.MkdirAll(confDir, 0755)
	}
	if ok, _ := tools.Exists("tor"); !ok {
		logger.Fatal(fmt.Errorf("fatal error: Tor is not installed"))
	}
	if ok, _ := tools.Exists("NetworkManager"); !ok {
		logger.Fatal("fatal error: NetworkManager is not installed")
	}
	if _, err := os.Stat("/run/tor"); os.IsNotExist(err) {
		os.Mkdir("/run/tor", 0700)
	}
	if _, err := os.Stat("/var/lib/tor/hidemego"); os.IsNotExist(err) {
		os.Mkdir("/var/lib/tor/hidemego", 0700)
	}
	if ok, _ := tools.Exists("setenforce"); ok {
		if err := tor.ChangeDirOwner("/run/tor"); err != nil {
			logger.Fatal("Can't Change Dir Owner on /run/tor", err)
		}
		if err := tor.ChangeDirOwner("/var/lib/tor"); err != nil {
			logger.Fatal("Can't Change Dir Owner on /var/lib/tor", err)
		}
		if err := tor.ChangeDirOwner("/var/run/tor"); err != nil {
			logger.Fatal("Can't Change Dir Owner on /var/run/tor", err)
		}
		if err := tor.ChangeDirOwner("/var/lib/tor/hidemego"); err != nil {
			logger.Fatal("Can't Change Dir Owner on /var/lib/tor/hidemego", err)
		}
		if !linux.HasSELPort(socksDestPort) {
			logger.Println(fmt.Sprintf("Opening TCP Sock Dest Port on %d", socksDestPort))
			if err := linux.SELManage(socksDestPort, true); err != nil {
				logger.Fatal(err)
			} // socks dest
		}
		if !linux.HasSELPort(socksAuthPort) {
			logger.Println(fmt.Sprintf("Opening TCP Sock Auth Port on %d", socksAuthPort))
			if err := linux.SELManage(socksAuthPort, true); err != nil {
				logger.Fatal(err)
			} // socks auth
		}
		if !linux.HasSELPort(controlPort) {
			logger.Println(fmt.Sprintf("Opening TCP Control Port on %d", controlPort))
			if err := linux.SELManage(controlPort, true); err != nil {
				logger.Fatal(err)
			} // control
		}
		if !linux.HasSELPort(torPort) {
			logger.Println(fmt.Sprintf("Opening TCP Port on %d", torPort))
			if err := linux.SELManage(torPort, true); err != nil {
				logger.Fatal(err)
			} // transport 
		}
		if !linux.HasSELPort(dnsPort) {
			logger.Println(fmt.Sprintf("Opening TCP and UDP DNS Ports on %d", dnsPort))
			if err := linux.SELManage(dnsPort, true, true); err != nil {
				logger.Fatal(err)
			} // dns
		}
	}
	if !nokch {
		logger.Println("Saving Kernel Configuration")
		if err := linux.SaveKernelConfigs(); err != nil {
			logger.Println("Can't Save Kernel Configuration. Restart The System after `stop`")
			os.RemoveAll(confDir)
			logger.Println(err)
		}
		logger.Println("Securing Kernel")
		linux.PrepareLinuxKernel()

	}
}

func close() {
	args, err := tools.PreviousArgs()
	if err != nil {
		logger.Fatal("Can't get Parsed Flags")
	}
	if len(args) > 0 {
		fl.Parse(args)
	}

	logger.Println("Stopping Tor Service")
	if err := tor.Stop(); err != nil {
		logger.Fatal("Can't Stop Tor:", err)
	}
	logger.Println("Restoring resolv.conf")
	_ = linux.RestoreResolvConf()

	if ifaces != "" {
		var ifa []string = []string{}
		tifa := strings.Split(ifaces, ",")
		for _, t := range tifa {
			if t != "" {
				ifa = append(ifa, t)
			}
		}
		for _, r := range ifa {
			logger.Println("Changing MAC Address for", ifaces)

			if !linux.HasIface(r) {
				logger.Println(fmt.Sprintf("Invalid Network Interface %s Found! Skipping", r))
				continue
			}

			if err := linux.IPSet(r, "down"); err != nil {
				logger.Fatal(err)
			}
			dmac, err := linux.DefaultMacAddr(r)
			if err != nil {
				logger.Fatal(fmt.Sprintf("Can't Restore Default MAC Address for %s", r))
			}
			logger.Println(fmt.Sprintf("Restoring %s MAC Address: %s", r, dmac))
			if err := linux.IPSetMACAddr(r, dmac); err != nil {
				logger.Fatal(err)
			}
			if err := linux.IPSet(r, "up"); err != nil {
				logger.Fatal(err)
			}
			time.Sleep(3 * time.Second)
		}
	}

	logger.Println("Removing Hidemego TorRC File")
	if err := tor.RemoveTorRc(); err != nil {
		logger.Fatal("Can't Remove hidemego.torrc", err)
	}
	logger.Println("Removing Hidemego Directory")
	if err := tor.RemoveHideMeGoDir(); err != nil {
		logger.Fatal("Can't Remove Hidemego Dir:", err)
	}
	if linux.HasSELPort(socksDestPort) && socksDestPort != 9051 {
		logger.Println(fmt.Sprintf("Closing TCP Port on %d", socksDestPort))
		if err := linux.SELManage(socksDestPort, false); err != nil {
			logger.Fatal(fmt.Sprintf("Can't Close Port: %d ", socksDestPort), err)
		} // socks dest
	}
	if linux.HasSELPort(socksAuthPort) {
		logger.Println(fmt.Sprintf("Closing TCP Port on %d", socksAuthPort))
		if err := linux.SELManage(socksAuthPort, false); err != nil {
			logger.Fatal(fmt.Sprintf("Can't Close Port: %d ", socksAuthPort), err)
		} // socks auth
	}
	if linux.HasSELPort(controlPort) {
		logger.Println(fmt.Sprintf("Closing TCP Port on %d", controlPort))
		if err := linux.SELManage(controlPort, false); err != nil {
			logger.Fatal(fmt.Sprintf("Can't Close Port: %d ", controlPort), err)
		} // control
	}
	if linux.HasSELPort(torPort) {
		logger.Println(fmt.Sprintf("Closing TCP Port on %d", torPort))
		if err := linux.SELManage(torPort, false); err != nil {
			logger.Fatal(fmt.Sprintf("Can't Close Port: %d ", torPort), err)
		} // transport
	}
	if linux.HasSELPort(dnsPort) {
		logger.Println(fmt.Sprintf("Closing TCP and UDP Ports on %d", dnsPort))
		_ = linux.SELManage(dnsPort, false, true)
	}

	logger.Println("Flushing IPTables Rules")
	if err := linux.FlushIPTablesRules(); err != nil {
		logger.Fatal("Can't Flush IPTables Rules:", err)
	}
	if !nokch {
		logger.Println("Restoring Kernel Configuration...")
		if err := linux.RestoreKernelConfig(); err != nil {
			logger.Println("Can't Restore Kernel Configuration:", err)
			logger.Fatal("Restart Your System to Revert Some Changes")
		}
	}

	logger.Println("Restarting the network using NetworkManager")
	if err := linux.RestartNetwork(true); err != nil {
		logger.Fatal("Can't Restart the Network:", err)
	}
	time.Sleep(3 * time.Second)
	if tools.CheckConn() {
		ip, err := tools.GetIPAddress()
		if err != nil {
			logger.Fatal("Can't Get IP Address:", err)
		}
		logger.Println("Your new IP is", ip)
	}
	logger.Println("Removing Hidemego Config Directory")
	os.RemoveAll(confDir)
}

func main() {

	fl.StringVar(&torPass, "pass", "", "The Tor Control Authentication Password")
	fl.StringVar(&torUser, "user", "", "Tor process user name. If no value is passed. Hidemego will parse defaults-torrc to identify user")
	fl.IntVar(&torPort, "tport", 9040, "Tor Port")
	fl.IntVar(&socksDestPort, "sdport", 9051, "Socks Destination Port for 127.0.0.1.1")
	fl.IntVar(&socksAuthPort, "saport", 9151, "Socks Authentication Port for 127.0.0.1.0")
	fl.IntVar(&controlPort, "cport", 9052, "Tor Control Port for 127.0.0.1")
	fl.IntVar(&dnsPort, "dport", 5354, "DNS Port")
	fl.IntVar(&torID, "id", 0, "Tor user id. If no value is passed Hidemego will parse default-torrc to identify user and related id")
	fl.StringVar(&ifaces, "ifaces", "", "Interfaces that must change MAC Address (separed by comma if multiple interfaces)")
	fl.BoolVar(&no5, "no5", false, "Excludes Nodes from 5 eyes countries")
	fl.BoolVar(&no9, "no9", false, "Excludes Nodes from 9 eyes countries")
	fl.BoolVar(&no14, "no14", false, "Excludes Nodes from 14 eyes countries")
	fl.BoolVar(&no14p, "no14p", false, "Excludes Nodes from 14 eyes countries plus other dangerous countries")
	fl.BoolVar(&nokch, "nkc", false, "Don't Change Kernel Configuration using Sysctl")

	if len(os.Args) < 2 {
		fl.Usage()
		return
	}
	args := os.Args[1:]

	command := args[0]

	sigs := make(chan os.Signal, 2)

	if !tools.IsRoot() {
		logger.Fatal(fmt.Errorf("You MUST BE ROOT to run this software"))
	}

	switch command {
	case "start":

		fl.Parse(args[1:])

		if len(args) > 1 {
			tools.SaveFlags(args[1:])
		}

		signal.Notify(sigs, os.Interrupt)
		go func() {
			// clean system on interrupt
			sig := <-sigs
			logger.Println("Caught:", sig)
			once.Do(close)
			os.Exit(1)
		}()

		logger.Println("Starting Hidemego Service to Anonymize the System")
		initialize()
		nontor := tor.NonTor()
		var countries string = tor.Countries

		var err error
		if torID == 0 {
			logger.Println("Detecting Tor ID...")
			if torID, err = tor.ID(); err != nil {
				logger.Fatal("Can't Automatically Detect Tor ID:", err)
			}
			logger.Println(fmt.Sprintf("Tor ID found: %d", torID))
		}

		if torUser == "" {
			logger.Println("Detecting Tor User...")
			torUser, err = tor.Usr()
			if err != nil {
				logger.Fatal("Can't Automatically Detect Tor User:", err)
			}
			logger.Println(fmt.Sprintf("Tor User found: %s", torUser))
		}

		if no5 {
			logger.Println("Skipping node families from 5 eyes countries")
			countries = tor.NoEyes(tor.Eyes5C)
		} else if no9 {
			logger.Println("Skipping node families from 9 eyes countries")
			countries = tor.NoEyes(tor.Eyes9C)
		} else if no14 {
			logger.Println("Skipping node families from 14 eyes countries")
			countries = tor.NoEyes(tor.Eyes14C)
		} else if no14p {
			logger.Println("Skipping node families from 14 eyes countries and others bad countries")
			countries = tor.NoEyes(tor.Eyes14CPlus)
		}

		logger.Println("Setting up Hidemego TorRC...")
		if torPass != "" {
			if err := tor.SetTorRC(countries, torPort, socksDestPort, socksAuthPort, controlPort, dnsPort, torUser, torPass); err != nil {
				logger.Fatal("Can't Setup Hidemego TorRC")
			}
		} else {
			if err := tor.SetTorRC(countries, torPort, socksDestPort, socksAuthPort, controlPort, dnsPort, torUser); err != nil {
				logger.Fatal("Can't Setup Hidemego TorRC")
			}
		}

		if ifaces != "" {
			var ifa []string = []string{}

			logger.Println("Changing MAC Address for", ifaces)
			tifa := strings.Split(ifaces, ",")
			for _, t := range tifa {
				if t != "" {
					ifa = append(ifa, t)
				}
			}
			for _, r := range ifa {
				if !linux.HasIface(r) {
					logger.Println(fmt.Sprintf("Invalid Network Interface %s Found! Skipping", r))
					continue
				}
				logger.Println("Generating MAC Address for", r)
				mac, err := tools.RandMACAddr()
				if err != nil {
					logger.Fatal("Could not Generate Random MAC Address for", r, err)
				}
				logger.Println("Assigning", mac, "to", r)
				if err := linux.IPSet(r, "down"); err != nil {
					logger.Fatal(err)
				}
				if err := linux.IPSetMACAddr(r, mac); err != nil {
					logger.Fatal(err)
				}
				if err := linux.IPSet(r, "up"); err != nil {
					logger.Fatal(err)
				}
			}
		}
		logger.Println("Changing resolv.conf...")
		if err := linux.SetResolvConf(); err != nil {
			logger.Fatal("Can't Change resolv.conf:", err)
		}
		logger.Println("Restarting Tor Service")
		if tor.Restart(); err != nil {
			logger.Fatal("Can't Restart Tor Service:", err)
		}
		logger.Println("Setting Up IPTables Rules")
		if err := linux.SetIPTablesRules(nontor, torID, torPort, dnsPort); err != nil {
			logger.Fatal("Can't Setup IPTables Rules", err)
		}
		time.Sleep(3 * time.Second)
		if tools.CheckConn() {
			ip, err := tools.GetIPAddress()
			if err != nil {
				logger.Fatal("Could Not Get IP Address:", err)
			}
			logger.Println("Your new IP Address is", ip)
		}
	case "new":
		logger.Println("Changing Your Identity")
		ip, err := tor.ChangeIdentity(torPass, torPort)
		if err != nil {
			logger.Fatal("Can't Change Your Identity:", err)
		}
		logger.Println("Your New IP Address is", ip)
	case "stop":
		close()
	default:
		fl.Usage()
	}
}

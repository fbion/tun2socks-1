package configure

import (
	"errors"
	"log"
	"net/url"

	"gopkg.in/gcfg.v1"
)

const (
	DnsDefaultPort         = 53
	DnsDefaultTtl          = 600
	DnsDefaultPacketSize   = 4096
	DnsDefaultReadTimeout  = 5
	DnsDefaultWriteTimeout = 5
	DnsIPPoolMaxSpace      = 0x3ffff // 4*65535
)

// GeneralConfig ini
type GeneralConfig struct {
	Network string // tun network
	Mtu     uint32
}

// PprofConfig ini
type PprofConfig struct {
	Enabled  bool
	ProfHost string `gcfg:"prof-host"`
	ProfPort uint16 `gcfg:"prof-port"`
}

// DnsConfig ini
type DnsConfig struct {
	DnsMode         string   `gcfg:"dns-mode"`
	DnsPort         uint16   `gcfg:"dns-port"`
	DnsTtl          uint     `gcfg:"dns-ttl"`
	DnsPacketSize   uint16   `gcfg:"dns-packet-size"`
	DnsReadTimeout  uint     `gcfg:"dns-read-timeout"`
	DnsWriteTimeout uint     `gcfg:"dns-write-timeout"`
	Nameserver      []string // backend dns
}

type RouteConfig struct {
	V []string
}

type PatternConfig struct {
	Proxy  string
	Scheme string
	V      []string
}

type RuleConfig struct {
	Pattern []string
	Final   string
}

type ProxyConfig struct {
	Url     string
	Default bool
}

type UdpConfig struct {
	Proxy   string
	Enabled bool
	Timeout int
}

type AppConfig struct {
	General GeneralConfig
	Pprof   PprofConfig
	Dns     DnsConfig
	Udp     UdpConfig
	Route   RouteConfig
	Proxy   map[string]*ProxyConfig
	Pattern map[string]*PatternConfig
	Rule    RuleConfig
	File    string
}

func (cfg *AppConfig) check() error {
	// TODO
	return nil
}

// Parse the config.ini file to AppConfig
func (cfg *AppConfig) Parse(filename string) error {
	// set default value
	cfg.General.Network = "198.18.0.0/15"
	cfg.General.Mtu = 1500

	cfg.Pprof.Enabled = true
	cfg.Pprof.ProfHost = "127.0.0.1"
	cfg.Pprof.ProfPort = 6060

	cfg.Dns.DnsMode = "fake"
	cfg.Dns.DnsPort = DnsDefaultPort
	cfg.Dns.DnsTtl = DnsDefaultTtl
	cfg.Dns.DnsPacketSize = DnsDefaultPacketSize
	cfg.Dns.DnsReadTimeout = DnsDefaultReadTimeout
	cfg.Dns.DnsWriteTimeout = DnsDefaultWriteTimeout

	cfg.Udp.Enabled = true
	cfg.Udp.Timeout = 300

	// decode config value
	err := gcfg.ReadFileInto(cfg, filename)
	if err != nil {
		return err
	}

	// set backend dns default value
	if len(cfg.Dns.Nameserver) == 0 {
		cfg.Dns.Nameserver = append(cfg.Dns.Nameserver, "114.114.114.114:53")
		cfg.Dns.Nameserver = append(cfg.Dns.Nameserver, "223.5.5.5:53")
	}

	err = cfg.check()
	if err != nil {
		return err
	}

	cfg.File = filename
	return nil
}

// GetProxy addr from name
func (cfg *AppConfig) GetProxy(name string) string {
	proxyConfig := cfg.Proxy[name]
	url, err := url.Parse(proxyConfig.Url)
	if err != nil {
		log.Println("Parse url failed", err)
		return ""
	}
	return url.Host
}

// DefaultPorxy return default proxy addr, eg: socks5://127.0.0.1:1080, return 127.0.0.1:1080
func (cfg *AppConfig) DefaultPorxy() (string, error) {
	proxyConfig := cfg.DefaultPorxyConfig()
	u, err := url.Parse(proxyConfig.Url)
	if err != nil {
		log.Println("Parse url failed", err)
		return "", err
	}
	return u.Host, nil
}

// DefaultPorxyConfig return the default ProxyConfig pointer
func (cfg *AppConfig) DefaultPorxyConfig() *ProxyConfig {
	for _, proxyConfig := range cfg.Proxy {
		if proxyConfig.Default {
			return proxyConfig
		}
	}
	return nil
}

// UdpProxy return the configed udp proxy
func (cfg *AppConfig) UdpProxy() (string, error) {
	proxyConfig := cfg.Proxy[cfg.Udp.Proxy]
	if proxyConfig == nil {
		proxyConfig = cfg.DefaultPorxyConfig()
	}
	if proxyConfig != nil {
		u, err := url.Parse(proxyConfig.Url)
		if err != nil {
			log.Println("Parse url failed", err)
			return "", err
		}
		return u.Host, nil
	}

	return "", errors.New("404")
}

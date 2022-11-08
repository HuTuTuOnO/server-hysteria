package app

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

const (
	mbpsToBps   = 125000
	minSpeedBPS = 16384

	DefaultStreamReceiveWindow     = 15728640 // 15 MB/s
	DefaultConnectionReceiveWindow = 67108864 // 64 MB/s
	DefaultMaxIncomingStreams      = 1024

	DefaultALPN = "h3"

	ServerMaxIdleTimeoutSec = 60
)

var rateStringRegexp = regexp.MustCompile(`^(\d+)\s*([KMGT]?)([Bb])ps$`)

type ServerConfig struct {
	Listen   string `json:"listen"`
	Protocol string `json:"protocol"`
	CertFile string `json:"cert"`
	KeyFile  string `json:"key"`
	// Optional below
	UpMbps              int    `json:"up_mbps"`
	DownMbps            int    `json:"down_mbps"`
	DisableUDP          bool   `json:"disable_udp"`
	Obfs                string `json:"obfs"`
	ALPN                string `json:"alpn"`
	ReceiveWindowConn   uint64 `json:"recv_window_conn"`
	ReceiveWindowClient uint64 `json:"recv_window_client"`
	MaxConnClient       int    `json:"max_conn_client"`
	DisableMTUDiscovery bool   `json:"disable_mtu_discovery"`
}

func (c *ServerConfig) Speed() (uint64, uint64, error) {
	var up, down uint64
	up = uint64(c.UpMbps) * mbpsToBps
	down = uint64(c.DownMbps) * mbpsToBps
	return up, down, nil
}

func (c *ServerConfig) Check() error {
	if len(c.Listen) == 0 {
		return errors.New("missing listen address")
	}
	if up, down, err := c.Speed(); err != nil || (up != 0 && up < minSpeedBPS) || (down != 0 && down < minSpeedBPS) {
		return errors.New("invalid speed")
	}
	if (c.ReceiveWindowConn != 0 && c.ReceiveWindowConn < 65536) ||
		(c.ReceiveWindowClient != 0 && c.ReceiveWindowClient < 65536) {
		return errors.New("invalid receive window size")
	}
	if c.MaxConnClient < 0 {
		return errors.New("invalid max connections per client")
	}
	return nil
}

func (c *ServerConfig) Fill() {
	if len(c.ALPN) == 0 {
		c.ALPN = DefaultALPN
	}
	if c.ReceiveWindowConn == 0 {
		c.ReceiveWindowConn = DefaultStreamReceiveWindow
	}
	if c.ReceiveWindowClient == 0 {
		c.ReceiveWindowClient = DefaultConnectionReceiveWindow
	}
	if c.MaxConnClient == 0 {
		c.MaxConnClient = DefaultMaxIncomingStreams
	}
}

func (c *ServerConfig) String() string {
	return fmt.Sprintf("%+v", *c)
}

func stringToBps(s string) uint64 {
	if s == "" {
		return 0
	}
	m := rateStringRegexp.FindStringSubmatch(s)
	if m == nil {
		return 0
	}
	var n uint64
	switch m[2] {
	case "K":
		n = 1 << 10
	case "M":
		n = 1 << 20
	case "G":
		n = 1 << 30
	case "T":
		n = 1 << 40
	default:
		n = 1
	}
	v, _ := strconv.ParseUint(m[1], 10, 64)
	n = v * n
	if m[3] == "b" {
		// Bits, need to convert to bytes
		n = n >> 3
	}
	return n
}

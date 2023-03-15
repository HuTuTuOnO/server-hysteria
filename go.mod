module github.com/xflash-panda/server-hysteria

go 1.18

require (
	github.com/coreos/go-iptables v0.6.0
	github.com/fsnotify/fsnotify v1.6.0
	github.com/go-resty/resty/v2 v2.7.0
	github.com/google/gopacket v1.1.19
	github.com/lunixbochs/struc v0.0.0-20200707160740-784aaebc1d40
	github.com/quic-go/quic-go v0.33.0
	github.com/sirupsen/logrus v1.9.0
	github.com/txthinking/socks5 v0.0.0-20220212043548-414499347d4a
	github.com/urfave/cli/v2 v2.20.3
	github.com/xtls/xray-core v1.6.1
	golang.org/x/sys v0.5.0
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/quic-go/qtls-go1-19 v0.2.1 // indirect
	github.com/quic-go/qtls-go1-20 v0.1.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/txthinking/runnergroup v0.0.0-20210608031112-152c7c4432bf // indirect
	github.com/txthinking/x v0.0.0-20210326105829-476fab902fbe // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/crypto v0.4.0 // indirect
	golang.org/x/exp v0.0.0-20221205204356-47842c84f3db // indirect
	golang.org/x/net v0.4.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

replace github.com/quic-go/quic-go => github.com/apernet/quic-go v0.32.1-0.20230226201325-e07aae1a800b

module github.com/xflash-panda/server-hysteria

go 1.20

require (
	github.com/coreos/go-iptables v0.6.0
	github.com/fsnotify/fsnotify v1.6.0
	github.com/google/gopacket v1.1.19
	github.com/lunixbochs/struc v0.0.0-20200707160740-784aaebc1d40
	github.com/quic-go/quic-go v0.34.0
	github.com/sirupsen/logrus v1.9.3
	github.com/txthinking/socks5 v0.0.0-20220212043548-414499347d4a
	github.com/urfave/cli/v2 v2.20.3
	github.com/xflash-panda/server-client v0.0.6
	golang.org/x/sys v0.14.0
)

require (
	github.com/go-resty/resty/v2 v2.10.0 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/go-task/slim-sprig v0.0.0-20210107165309-348f09dbbbc0 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/google/pprof v0.0.0-20210407192527-94a9f03dee38 // indirect
	github.com/onsi/ginkgo/v2 v2.2.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/quic-go/qtls-go1-19 v0.3.2 // indirect
	github.com/quic-go/qtls-go1-20 v0.2.2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/txthinking/runnergroup v0.0.0-20210608031112-152c7c4432bf // indirect
	github.com/txthinking/x v0.0.0-20210326105829-476fab902fbe // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/crypto v0.15.0 // indirect
	golang.org/x/exp v0.0.0-20221205204356-47842c84f3db // indirect
	golang.org/x/mod v0.8.0 // indirect
	golang.org/x/net v0.18.0 // indirect
	golang.org/x/tools v0.6.0 // indirect; indirect// indirect
	google.golang.org/protobuf v1.28.2-0.20230118093459-a9481185b34d // indirect
)

replace github.com/quic-go/quic-go => github.com/apernet/quic-go v0.34.1-0.20230507231629-ec008b7e8473

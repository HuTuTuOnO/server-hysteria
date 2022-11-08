module github.com/xflash-panda/server-hysteria

go 1.18

require (
	github.com/coreos/go-iptables v0.6.0
	github.com/fsnotify/fsnotify v1.6.0
	github.com/go-resty/resty/v2 v2.7.0
	github.com/google/gopacket v1.1.19
	github.com/lucas-clemente/quic-go v0.30.0
	github.com/lunixbochs/struc v0.0.0-20200707160740-784aaebc1d40
	github.com/sirupsen/logrus v1.9.0
	github.com/txthinking/socks5 v0.0.0-20220212043548-414499347d4a
	github.com/urfave/cli/v2 v2.20.3
	github.com/xtls/xray-core v1.6.1
	golang.org/x/sys v0.1.0
)

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/francoispqt/gojay v1.2.13 // indirect
	github.com/go-task/slim-sprig v0.0.0-20210107165309-348f09dbbbc0 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/pprof v0.0.0-20210720184732-4bb14d4b1be1 // indirect
	github.com/klauspost/compress v1.15.11 // indirect
	github.com/marten-seemann/qtls-go1-18 v0.1.3 // indirect
	github.com/marten-seemann/qtls-go1-19 v0.1.1 // indirect
	github.com/onsi/ginkgo/v2 v2.2.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pires/go-proxyproto v0.6.2 // indirect
	github.com/refraction-networking/utls v1.1.5 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/txthinking/runnergroup v0.0.0-20210608031112-152c7c4432bf // indirect
	github.com/txthinking/x v0.0.0-20210326105829-476fab902fbe // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/crypto v0.1.0 // indirect
	golang.org/x/exp v0.0.0-20221019170559-20944726eadf // indirect
	golang.org/x/mod v0.6.0 // indirect
	golang.org/x/net v0.1.0 // indirect
	golang.org/x/tools v0.2.0 // indirect
	google.golang.org/grpc v1.50.1 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

replace github.com/lucas-clemente/quic-go => github.com/HyNetwork/quic-go v0.30.1-0.20221105180419-83715d7269a8

replace github.com/LiamHaworth/go-tproxy => github.com/HyNetwork/go-tproxy v0.0.0-20221025153553-ed04a2935f88

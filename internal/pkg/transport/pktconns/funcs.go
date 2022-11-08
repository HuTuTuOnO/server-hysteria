package pktconns

import (
	"net"

	"github.com/xflash-panda/server-hysteria/internal/pkg/transport/pktconns/faketcp"
	"github.com/xflash-panda/server-hysteria/internal/pkg/transport/pktconns/obfs"
	"github.com/xflash-panda/server-hysteria/internal/pkg/transport/pktconns/udp"
	"github.com/xflash-panda/server-hysteria/internal/pkg/transport/pktconns/wechat"
)

type (
	ClientPacketConnFunc func(server string) (net.PacketConn, net.Addr, error)
	ServerPacketConnFunc func(listen string) (net.PacketConn, error)
)

type (
	ClientPacketConnFuncFactory func(obfsPassword string) ClientPacketConnFunc
	ServerPacketConnFuncFactory func(obfsPassword string) ServerPacketConnFunc
)

func NewServerUDPConnFunc(obfsPassword string) ServerPacketConnFunc {
	if obfsPassword == "" {
		return func(listen string) (net.PacketConn, error) {
			laddrU, err := net.ResolveUDPAddr("udp", listen)
			if err != nil {
				return nil, err
			}
			return net.ListenUDP("udp", laddrU)
		}
	} else {
		return func(listen string) (net.PacketConn, error) {
			ob := obfs.NewXPlusObfuscator([]byte(obfsPassword))
			laddrU, err := net.ResolveUDPAddr("udp", listen)
			if err != nil {
				return nil, err
			}
			udpConn, err := net.ListenUDP("udp", laddrU)
			if err != nil {
				return nil, err
			}
			return udp.NewObfsUDPConn(udpConn, ob), nil
		}
	}
}

func NewServerWeChatConnFunc(obfsPassword string) ServerPacketConnFunc {
	if obfsPassword == "" {
		return func(listen string) (net.PacketConn, error) {
			laddrU, err := net.ResolveUDPAddr("udp", listen)
			if err != nil {
				return nil, err
			}
			udpConn, err := net.ListenUDP("udp", laddrU)
			if err != nil {
				return nil, err
			}
			return wechat.NewObfsWeChatUDPConn(udpConn, nil), nil
		}
	} else {
		return func(listen string) (net.PacketConn, error) {
			ob := obfs.NewXPlusObfuscator([]byte(obfsPassword))
			laddrU, err := net.ResolveUDPAddr("udp", listen)
			if err != nil {
				return nil, err
			}
			udpConn, err := net.ListenUDP("udp", laddrU)
			if err != nil {
				return nil, err
			}
			return wechat.NewObfsWeChatUDPConn(udpConn, ob), nil
		}
	}
}

func NewServerFakeTCPConnFunc(obfsPassword string) ServerPacketConnFunc {
	if obfsPassword == "" {
		return func(listen string) (net.PacketConn, error) {
			return faketcp.Listen("tcp", listen)
		}
	} else {
		return func(listen string) (net.PacketConn, error) {
			ob := obfs.NewXPlusObfuscator([]byte(obfsPassword))
			fakeTCPListener, err := faketcp.Listen("tcp", listen)
			if err != nil {
				return nil, err
			}
			return faketcp.NewObfsFakeTCPConn(fakeTCPListener, ob), nil
		}
	}
}

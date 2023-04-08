package core

import (
	"bytes"
	"context"
	"github.com/lunixbochs/struc"
	"github.com/quic-go/quic-go"
	"github.com/xflash-panda/server-hysteria/internal/app/service"
	"github.com/xflash-panda/server-hysteria/internal/pkg/transport"
	"github.com/xflash-panda/server-hysteria/internal/pkg/utils"
	"math/rand"
	"net"
	"strconv"
	"sync"
)

const udpBufferSize = 4096

type serverClient struct {
	CC               quic.Connection
	Transport        *transport.ServerTransport
	UserId           int
	DisableUDP       bool
	CTCPRequestFunc  TCPRequestFunc
	CTCPErrorFunc    TCPErrorFunc
	CUDPRequestFunc  UDPRequestFunc
	CUDPErrorFunc    UDPErrorFunc
	TrafficItem      *service.TrafficItem
	udpSessionMutex  sync.RWMutex
	udpSessionMap    map[uint32]transport.STPacketConn
	nextUDPSessionID uint32
	udpDefragger     defragger
}

func newServerClient(cc quic.Connection, tr *transport.ServerTransport, userId int, disableUDP bool,
	trafficItem *service.TrafficItem,
	CTCPRequestFunc TCPRequestFunc, CTCPErrorFunc TCPErrorFunc,
	CUDPRequestFunc UDPRequestFunc, CUDPErrorFunc UDPErrorFunc,
) *serverClient {
	sc := &serverClient{
		CC:              cc,
		Transport:       tr,
		UserId:          userId,
		DisableUDP:      disableUDP,
		TrafficItem:     trafficItem,
		CTCPRequestFunc: CTCPRequestFunc,
		CTCPErrorFunc:   CTCPErrorFunc,
		CUDPRequestFunc: CUDPRequestFunc,
		CUDPErrorFunc:   CUDPErrorFunc,
		udpSessionMap:   make(map[uint32]transport.STPacketConn),
	}
	return sc
}

func (c *serverClient) ClientAddr() net.Addr {
	// quic.Connection's remote address may change since we have connection migration now,
	// so logs need to dynamically get the remote address every time.
	return c.CC.RemoteAddr()
}

func (c *serverClient) Run() error {
	if !c.DisableUDP {
		go func() {
			for {
				msg, err := c.CC.ReceiveMessage()
				if err != nil {
					break
				}
				c.handleMessage(msg)
			}
		}()
	}
	for {
		stream, err := c.CC.AcceptStream(context.Background())
		if err != nil {
			return err
		}

		if c.TrafficItem != nil {
			c.TrafficItem.Count.Add(1)
		}

		go func() {
			stream := &qStream{stream}
			c.handleStream(stream)
			_ = stream.Close()
		}()
	}
}

func (c *serverClient) handleStream(stream quic.Stream) {
	// Read request
	var req clientRequest
	err := struc.Unpack(stream, &req)
	if err != nil {
		return
	}
	if !req.UDP {
		// TCP connection
		c.handleTCP(stream, req.Host, req.Port)
	} else if !c.DisableUDP {
		// UDP connection
		c.handleUDP(stream)
	} else {
		// UDP disabled
		_ = struc.Pack(stream, &serverResponse{
			OK:      false,
			Message: "UDP disabled",
		})
	}
}

func (c *serverClient) handleMessage(msg []byte) {
	var udpMsg udpMessage
	err := struc.Unpack(bytes.NewBuffer(msg), &udpMsg)
	if err != nil {
		return
	}
	dfMsg := c.udpDefragger.Feed(udpMsg)
	if dfMsg == nil {
		return
	}
	c.udpSessionMutex.RLock()
	conn, ok := c.udpSessionMap[dfMsg.SessionID]
	c.udpSessionMutex.RUnlock()
	if ok {
		// Session found, send the message
		var isDomain bool
		var ipAddr *net.IPAddr
		var err error

		ipAddr, isDomain, err = c.Transport.ResolveIPAddr(dfMsg.Host)
		if err != nil { // Special case for domain requests + SOCKS5 outbound
			return
		}

		addrEx := &transport.AddrEx{
			IPAddr: ipAddr,
			Port:   int(dfMsg.Port),
		}
		if isDomain {
			addrEx.Domain = dfMsg.Host
		}
		_, _ = conn.WriteTo(dfMsg.Data, addrEx)
		if c.TrafficItem != nil {
			c.TrafficItem.Up.Add(int64(len(dfMsg.Data)))
		}
	}

}

func (c *serverClient) handleTCP(stream quic.Stream, host string, port uint16) {
	addrStr := net.JoinHostPort(host, strconv.Itoa(int(port)))
	var isDomain bool
	var ipAddr *net.IPAddr
	var err error

	ipAddr, isDomain, err = c.Transport.ResolveIPAddr(host)

	if err != nil && !(isDomain && c.Transport.ProxyEnabled()) { // Special case for domain requests + SOCKS5 outbound
		_ = struc.Pack(stream, &serverResponse{
			OK:      false,
			Message: "host resolution failure",
		})
		c.CTCPErrorFunc(c.ClientAddr(), c.UserId, addrStr, err)
		return
	}
	c.CTCPRequestFunc(c.ClientAddr(), c.UserId, addrStr)

	var conn net.Conn // Connection to be piped

	addrEx := &transport.AddrEx{
		IPAddr: ipAddr,
		Port:   int(port),
	}
	if isDomain {
		addrEx.Domain = host
	}
	conn, err = c.Transport.DialTCP(addrEx)
	if err != nil {
		_ = struc.Pack(stream, &serverResponse{
			OK:      false,
			Message: err.Error(),
		})
		c.CTCPErrorFunc(c.ClientAddr(), c.UserId, addrStr, err)
		return
	}

	// So far so good if we reach here
	defer conn.Close()
	err = struc.Pack(stream, &serverResponse{
		OK: true,
	})
	if err != nil {
		return
	}
	if c.TrafficItem != nil {
		err = utils.Pipe2Way(stream, conn, func(i int) {
			if i > 0 {
				c.TrafficItem.Up.Add(int64(i))
			} else {
				c.TrafficItem.Down.Add(int64(-i))
			}
		})
	} else {
		err = utils.Pipe2Way(stream, conn, nil)
	}
	c.CTCPErrorFunc(c.ClientAddr(), c.UserId, addrStr, err)
}

func (c *serverClient) handleUDP(stream quic.Stream) {
	// Like in SOCKS5, the stream here is only used to maintain the UDP session. No need to read anything from it
	conn, err := c.Transport.ListenUDP()
	if err != nil {
		_ = struc.Pack(stream, &serverResponse{
			OK:      false,
			Message: "UDP initialization failed",
		})
		c.CUDPErrorFunc(c.ClientAddr(), c.UserId, 0, err)
		return
	}
	defer conn.Close()

	var id uint32
	c.udpSessionMutex.Lock()
	id = c.nextUDPSessionID
	c.udpSessionMap[id] = conn
	c.nextUDPSessionID += 1
	c.udpSessionMutex.Unlock()

	err = struc.Pack(stream, &serverResponse{
		OK:           true,
		UDPSessionID: id,
	})
	if err != nil {
		return
	}
	c.CUDPRequestFunc(c.ClientAddr(), c.UserId, id)

	// Receive UDP packets, send them to the client
	go func() {
		buf := make([]byte, udpBufferSize)
		for {
			n, rAddr, err := conn.ReadFrom(buf)
			if n > 0 {
				var msgBuf bytes.Buffer
				msg := udpMessage{
					SessionID: id,
					Host:      rAddr.IP.String(),
					Port:      uint16(rAddr.Port),
					FragCount: 1,
					Data:      buf[:n],
				}
				// try no frag first
				_ = struc.Pack(&msgBuf, &msg)
				sendErr := c.CC.SendMessage(msgBuf.Bytes())
				if sendErr != nil {
					if errSize, ok := sendErr.(quic.ErrMessageTooLarge); ok {
						// need to frag
						msg.MsgID = uint16(rand.Intn(0xFFFF)) + 1 // msgID must be > 0 when fragCount > 1
						fragMessages := fragUDPMessage(msg, int(errSize))
						for _, fragMsg := range fragMessages {
							msgBuf.Reset()
							_ = struc.Pack(&msgBuf, &fragMsg)
							_ = c.CC.SendMessage(msgBuf.Bytes())
						}
					}
				}
				if c.TrafficItem != nil {
					c.TrafficItem.Down.Add(int64(n))
				}
			}
			if err != nil {
				break
			}
		}
		_ = stream.Close()
	}()

	// Hold the stream until it's closed by the client
	buf := make([]byte, 1024)
	for {
		_, err = stream.Read(buf)
		if err != nil {
			break
		}
	}
	c.CUDPErrorFunc(c.ClientAddr(), c.UserId, id, err)

	// Remove the session
	c.udpSessionMutex.Lock()
	delete(c.udpSessionMap, id)
	c.udpSessionMutex.Unlock()
}

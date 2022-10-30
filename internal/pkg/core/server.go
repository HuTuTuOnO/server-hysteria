package core

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"

	"github.com/xflash-panda/server-hysteria/internal/pkg/congestion"

	"github.com/lucas-clemente/quic-go"
	"github.com/lunixbochs/struc"
	"github.com/xflash-panda/server-hysteria/internal/app/service"
	"github.com/xflash-panda/server-hysteria/internal/pkg/pmtud"
	"github.com/xflash-panda/server-hysteria/internal/pkg/transport"
)

type (
	ConnectFunc    func(addr net.Addr, auth []byte, sSend uint64, sRecv uint64) (bool, int)
	DisconnectFunc func(addr net.Addr, userId int, err error)
	TCPRequestFunc func(addr net.Addr, userId int, reqAddr string)
	TCPErrorFunc   func(addr net.Addr, userId int, reqAddr string, err error)
	UDPRequestFunc func(addr net.Addr, userId int, sessionID uint32)
	UDPErrorFunc   func(addr net.Addr, userId int, sessionID uint32, err error)
)

type Server struct {
	transport        *transport.ServerTransport
	sendBPS, recvBPS uint64
	disableUDP       bool

	connectFunc    ConnectFunc
	disconnectFunc DisconnectFunc
	tcpRequestFunc TCPRequestFunc
	tcpErrorFunc   TCPErrorFunc
	udpRequestFunc UDPRequestFunc
	udpErrorFunc   UDPErrorFunc
	userService    *service.UsersService

	pktConn  net.PacketConn
	listener quic.Listener
}

func NewServer(tlsConfig *tls.Config, quicConfig *quic.Config,
	pktConn net.PacketConn, transport *transport.ServerTransport,
	sendBPS uint64, recvBPS uint64, disableUDP bool, userService *service.UsersService,
	connectFunc ConnectFunc, disconnectFunc DisconnectFunc,
	tcpRequestFunc TCPRequestFunc, tcpErrorFunc TCPErrorFunc,
	udpRequestFunc UDPRequestFunc, udpErrorFunc UDPErrorFunc,
) (*Server, error) {
	quicConfig.DisablePathMTUDiscovery = quicConfig.DisablePathMTUDiscovery || pmtud.DisablePathMTUDiscovery
	listener, err := quic.Listen(pktConn, tlsConfig, quicConfig)
	if err != nil {
		_ = pktConn.Close()
		return nil, err
	}
	s := &Server{
		pktConn:        pktConn,
		listener:       listener,
		transport:      transport,
		sendBPS:        sendBPS,
		recvBPS:        recvBPS,
		disableUDP:     disableUDP,
		userService:    userService,
		connectFunc:    connectFunc,
		disconnectFunc: disconnectFunc,
		tcpRequestFunc: tcpRequestFunc,
		tcpErrorFunc:   tcpErrorFunc,
		udpRequestFunc: udpRequestFunc,
		udpErrorFunc:   udpErrorFunc,
	}
	return s, nil
}

func (s *Server) Serve() error {
	for {
		cc, err := s.listener.Accept(context.Background())
		if err != nil {
			return err
		}
		go s.handleClient(cc)
	}
}

func (s *Server) Close() error {
	err := s.listener.Close()
	_ = s.pktConn.Close()
	return err
}

func (s *Server) handleClient(cc quic.Connection) {
	// Expect the client to create a control stream to send its own information
	ctx, ctxCancel := context.WithTimeout(context.Background(), protocolTimeout)
	stream, err := cc.AcceptStream(ctx)
	ctxCancel()
	if err != nil {
		_ = qErrorProtocol.Send(cc)
		return
	}
	// Handle the control stream
	userId, ok, err := s.handleControlStream(cc, stream)
	if err != nil {
		_ = qErrorProtocol.Send(cc)
		return
	}
	if !ok {
		_ = qErrorAuth.Send(cc)
		return
	}
	// Start accepting streams and messages
	sc := newServerClient(cc, s.transport, userId, s.disableUDP, s.userService.GetTrafficItem(userId),
		s.tcpRequestFunc, s.tcpErrorFunc, s.udpRequestFunc, s.udpErrorFunc)
	err = sc.Run()
	_ = qErrorGeneric.Send(cc)
	s.disconnectFunc(cc.RemoteAddr(), userId, err)
}

// Auth & negotiate speed
func (s *Server) handleControlStream(cc quic.Connection, stream quic.Stream) (int, bool, error) {
	// Check version
	vb := make([]byte, 1)
	_, err := stream.Read(vb)
	if err != nil {
		return -1, false, err
	}
	if vb[0] != protocolVersion {
		return -1, false, fmt.Errorf("unsupported protocol version %d, expecting %d", vb[0], protocolVersion)
	}
	// Parse client hello
	var ch clientHello
	err = struc.Unpack(stream, &ch)
	if err != nil {
		return -1, false, err
	}
	// Speed
	if ch.Rate.SendBPS == 0 || ch.Rate.RecvBPS == 0 {
		return -1, false, errors.New("invalid rate from client")
	}
	serverSendBPS, serverRecvBPS := ch.Rate.RecvBPS, ch.Rate.SendBPS
	if s.sendBPS > 0 && serverSendBPS > s.sendBPS {
		serverSendBPS = s.sendBPS
	}
	if s.recvBPS > 0 && serverRecvBPS > s.recvBPS {
		serverRecvBPS = s.recvBPS
	}
	// Auth
	ok, userId := s.connectFunc(cc.RemoteAddr(), ch.Auth, serverSendBPS, serverRecvBPS)
	// Response
	err = struc.Pack(stream, &serverHello{
		OK: ok,
		Rate: maxRate{
			SendBPS: serverSendBPS,
			RecvBPS: serverRecvBPS,
		},
		Message: "Welcome",
	})
	if err != nil {
		return -1, false, err
	}
	// Set the congestion accordingly
	if ok {
		cc.SetCongestionControl(congestion.NewBrutalSender(serverSendBPS))
	}
	return userId, ok, nil
}

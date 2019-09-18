package gotcp

import (
	"net"
	"sync"
	"time"
)

type Config struct {
	PacketSendChanLimit    uint32 //chan发送的数据包的最大值 the limit of packet send channel
	PacketReceiveChanLimit uint32 //chan接收的数据包最大值   the limit of packet receive channel
}

type Server struct {
	config    *Config         // server的配置(发送接受最大值) 	server configuration
	callback  ConnCallback    // 信息回调					message callbacks in connection
	protocol  Protocol        // 用户数据包协议				customize packet protocol
	exitChan  chan struct{}   // 退出channel					notify all goroutines to shutdown
	waitGroup *sync.WaitGroup // 等待所有的goroutine 		wait for all goroutines
}

// 创建一个服务
// NewServer creates a server
func NewServer(config *Config, callback ConnCallback, protocol Protocol) *Server {
	return &Server{
		config:    config,
		callback:  callback,
		protocol:  protocol,
		exitChan:  make(chan struct{}),
		waitGroup: &sync.WaitGroup{},
	}
}

// 开启服务
// Start starts service
func (s *Server) Start(listener *net.TCPListener, acceptTimeout time.Duration) {
	s.waitGroup.Add(1)
	defer func() {
		listener.Close()
		s.waitGroup.Done()
	}()

	for {
		select {
		case <-s.exitChan:
			return

		default:
		}

		listener.SetDeadline(time.Now().Add(acceptTimeout))	//设置监听到期时间

		conn, err := listener.AcceptTCP()
		if err != nil {
			continue
		}

		s.waitGroup.Add(1)
		go func() {
			newConn(conn, s).Do()
			s.waitGroup.Done()
		}()
	}
}

// Stop stops service
func (s *Server) Stop() {
	close(s.exitChan)
	s.waitGroup.Wait()
}
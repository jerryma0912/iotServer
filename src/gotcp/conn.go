package gotcp

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// 定义错误的类型
var (
	ErrConnClosing   = errors.New("use of closed network connection")
	ErrWriteBlocking = errors.New("write packet was blocking")
	ErrReadBlocking  = errors.New("read packet was blocking")
)

// 一组接口，用于连接上的回调
// ConnCallback is an interface of methods that are used as callbacks on a connection
type ConnCallback interface {
	// 当连接被接受时被调用
	// OnConnect is called when the connection was accepted,
	// If the return value of false is closed
	OnConnect(*Conn) bool
	// 当收到包时被调用
	// OnMessage is called when the connection receives a packet,
	// If the return value of false is closed
	OnMessage(*Conn, Packet) bool
	//当关闭连接时被调用
	// OnClose is called when the connection closed
	OnClose(*Conn)
}

// Conn连接上发生的各种事件的一组回调，是原始conn的包装器
// Conn exposes a set of callbacks for the various events that occur on a connection
type Conn struct {
	srv               *Server		// conn所属的服务对象
	conn              *net.TCPConn  // 原始连接 		the raw connection
	extraData         interface{}   // 保存额外数据 	to save extra data
	closeOnce         sync.Once     // 关闭连接 		close the conn, once, per instance
	closeFlag         int32         // 关闭flag 		close flag
	closeChan         chan struct{} // 关闭channel	close chanel
	packetSendChan    chan Packet   // 包发送channel	packet send chanel
	packetReceiveChan chan Packet   // 包接收channel	packet receive chanel
}



// 创建一个conn的包装器
// newConn returns a wrapper of raw conn
func newConn(conn *net.TCPConn, srv *Server) *Conn {
	return &Conn{
		srv:               srv,
		conn:              conn,
		closeChan:         make(chan struct{}),
		packetSendChan:    make(chan Packet, srv.config.PacketSendChanLimit),
		packetReceiveChan: make(chan Packet, srv.config.PacketReceiveChanLimit),
	}
}

// 获取额外的数据
// GetExtraData gets the extra data from the Conn
func (c *Conn) GetExtraData() interface{} {
	return c.extraData
}

// 放入额外的数据
// PutExtraData puts the extra data with the Conn
func (c *Conn) PutExtraData(data interface{}) {
	c.extraData = data
}

// 返回原始的conn
// GetRawConn returns the raw net.TCPConn from the Conn
func (c *Conn) GetRawConn() *net.TCPConn {
	return c.conn
}

// 关闭连接
// Close closes the connection
func (c *Conn) Close() {
	c.closeOnce.Do(func() {
		atomic.StoreInt32(&c.closeFlag, 1)
		close(c.closeChan)
		close(c.packetSendChan)
		close(c.packetReceiveChan)
		c.conn.Close()
		c.srv.callback.OnClose(c)
	})
}

// 连接是否已关闭
// IsClosed indicates whether or not the connection is closed
func (c *Conn) IsClosed() bool {
	return atomic.LoadInt32(&c.closeFlag) == 1
}

// AsyncWritePacket异步写入数据包，此方法永远不会阻塞
// AsyncWritePacket async writes a packet, this method will never block
func (c *Conn) AsyncWritePacket(p Packet, timeout time.Duration) (err error) {
	if c.IsClosed() {
		return ErrConnClosing
	}

	defer func() {
		if e := recover(); e != nil {
			err = ErrConnClosing
		}
	}()

	if timeout == 0 {
		select {
		case c.packetSendChan <- p:	//写入数据包
			return nil

		default:
			return ErrWriteBlocking
		}

	} else {
		select {
		case c.packetSendChan <- p:	//写入数据包
			return nil

		case <-c.closeChan:
			return ErrConnClosing

		case <-time.After(timeout):	//超时时间
			return ErrWriteBlocking
		}
	}
}

//写入循环
func (c *Conn) writeLoop() {
	defer func() {
		recover()
		c.Close()
	}()

	for {
		select {
		case <-c.srv.exitChan:
			return

		case <-c.closeChan:
			return

		case p := <-c.packetSendChan:	//异步写入数据包	写入
			if c.IsClosed() {
				return
			}
			if _, err := c.conn.Write(p.Serialize()); err != nil {
				return
			}
		}
	}
}

//读循环
func (c *Conn) readLoop() {
	defer func() {
		recover()
		c.Close()
	}()

	for {
		select {
		case <-c.srv.exitChan:
			return

		case <-c.closeChan:
			return

		default:
		}

		p, err := c.srv.protocol.ReadPacket(c.conn)
		if err != nil {
			return
		}

		c.packetReceiveChan <- p	//packet写入receive channel
	}
}

//处理循环
func (c *Conn) handleLoop() {
	defer func() {
		recover()
		c.Close()
	}()

	for {
		select {
		case <-c.srv.exitChan:
			return

		case <-c.closeChan:
			return

		case p := <-c.packetReceiveChan:
			if c.IsClosed() {
				return
			}
			if !c.srv.callback.OnMessage(c, p) {	//接受的信息从receive channel中读取，并处理
				return
			}
		}
	}
}


// Do it
func (c *Conn) Do() {
	if !c.srv.callback.OnConnect(c) {
		return
	}

	asyncDo(c.handleLoop, c.srv.waitGroup)
	asyncDo(c.readLoop, c.srv.waitGroup)
	asyncDo(c.writeLoop, c.srv.waitGroup)
}


func asyncDo(fn func(), wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		fn()
		wg.Done()
	}()
}
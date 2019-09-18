package protocol

import (
	"IoTServer/src/gotcp"
	"bytes"
	"fmt"
	"net"
	"strings"
)

var (
	endTag = []byte("\r\n") // Telnet command's end tag
)

//--------Packet接口实现------------
// Packet结构
type TelnetPacket struct {
	pType string
	pData []byte
}

// 序列化数据包
func (p *TelnetPacket) Serialize() []byte {
	buf := p.pData
	buf = append(buf, endTag...)
	return buf
}

// 获取Packet类型名称
func (p *TelnetPacket) GetType() string {
	return p.pType
}

// 获取数据包
func (p *TelnetPacket) GetData() []byte {
	return p.pData
}

// 创建一个Packet对象
func NewTelnetPacket(pType string, pData []byte) *TelnetPacket {
	return &TelnetPacket{
		pType: pType,
		pData: pData,
	}
}

//--------Protocol接口实现------------
//Protocol结构体
type TelnetProtocol struct {
}

//
func (this *TelnetProtocol) ReadPacket(conn *net.TCPConn) (gotcp.Packet, error) {
	fullBuf := bytes.NewBuffer([]byte{})
	for {
		data := make([]byte, 1024)

		readLengh, err := conn.Read(data)

		if err != nil { //EOF, or worse
			return nil, err
		}

		if readLengh == 0 { // Connection maybe closed by the client
			return nil, gotcp.ErrConnClosing
		} else {
			fullBuf.Write(data[:readLengh])

			index := bytes.Index(fullBuf.Bytes(), endTag)
			if index > -1 {
				command := fullBuf.Next(index)
				fullBuf.Next(2)
				//fmt.Println(string(command))

				commandList := strings.Split(string(command), " ")
				if len(commandList) > 1 {
					return NewTelnetPacket(commandList[0], []byte(commandList[1])), nil
				} else {
					if commandList[0] == "quit" {
						return NewTelnetPacket("quit", command), nil
					} else {
						return NewTelnetPacket("unknow", command), nil
					}
				}
			}
		}
	}
}

//------------ConnCallback接口实现-------------
type TelnetCallback struct {
}

func (this *TelnetCallback) OnConnect(c *gotcp.Conn) bool {
	addr := c.GetRawConn().RemoteAddr()
	c.PutExtraData(addr)
	fmt.Println("OnConnect:", addr)
	c.AsyncWritePacket(NewTelnetPacket("unknow", []byte("Welcome to this Telnet Server")), 0)
	return true
}

func (this *TelnetCallback) OnMessage(c *gotcp.Conn, p gotcp.Packet) bool {
	packet := p.(*TelnetPacket)
	command := packet.GetData()
	commandType := packet.GetType()

	switch commandType {
	case "echo":
		c.AsyncWritePacket(NewTelnetPacket("echo", command), 0)
	case "login":
		c.AsyncWritePacket(NewTelnetPacket("login", []byte(string(command)+" has login")), 0)
	case "quit":
		return false
	default:
		c.AsyncWritePacket(NewTelnetPacket("unknow", []byte("unknow command")), 0)
	}

	return true
}

func (this *TelnetCallback) OnClose(c *gotcp.Conn) {
	fmt.Println("OnClose:", c.GetExtraData())
}
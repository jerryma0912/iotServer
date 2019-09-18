package archives

import (
	"IoTServer/src/gotcp"
	"log"
	"strings"
	"time"
)

type ConnArchive struct {

	ConnType  string      //连接类型
	ConnAddr  string      //连接的ip和端口
	LocPort   string      //本地连接端口
	Time      time.Time   //建立连接的时间
	Conn      *gotcp.Conn //连接对象
	ExtraData interface{} //额外信息（用于区分设备）
}

func NewconnArchive(conntype string, c *gotcp.Conn) *ConnArchive {

	remoteaddr := c.GetRawConn().RemoteAddr().String()
	localport := strings.Split(c.GetRawConn().LocalAddr().String(),":")[1]

	return &ConnArchive{
		Time:     time.Now(),
		Conn:     c,
		ConnType: conntype,
		ConnAddr: remoteaddr,
		LocPort:  localport,
	}
}

func (c *ConnArchive) GetConnect() *gotcp.Conn{

	return c.Conn
}

// 获取额外的数据
func (c *ConnArchive) GetExtraData() interface{} {
	return c.ExtraData
}

// 放入额外的数据
func (c *ConnArchive) PutExtraData(data interface{}) {
	c.ExtraData = data
}


func (c *ConnArchive) SendMessage(p gotcp.Packet) bool{

	err := c.Conn.AsyncWritePacket(p ,0)
	if err != nil{
		log.Println("SendMessage Err")
		return false
	}
	return true
}
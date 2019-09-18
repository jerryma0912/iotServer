package handle

import (
	"IoTServer/src/archives"
	"IoTServer/src/gotcp"
	"log"
	"net"
	"strings"
	"time"
)

//---------------------接口-------------------
type tcpHandle interface {
	UplinkParse(input []byte) (gotcp.Packet, error)
	UplinkHandle(packet gotcp.Packet)
	GetEquipId(packet gotcp.Packet) string
}

//------------------封装tcp服务-------------------
type TcpServer struct {
	networkType string
	address     string
	srv         *gotcp.Server
}

func NewTcpHandleProcess(nettype string, addr string, h tcpHandle) *TcpServer {

	//创建 tcp 配置
	config := &gotcp.Config{
		PacketSendChanLimit:    100,
		PacketReceiveChanLimit: 100,
	}
	//创建 tcp 服务
	handleProcess := &tcpHandleProcess{h}
	server := gotcp.NewServer(config, handleProcess, handleProcess)

	return &TcpServer{

		networkType: nettype,
		address:     addr,
		srv:         server,
	}
}

func (t *TcpServer) Start() {

	//创建 tcp 监听器
	tcpAddr, err := net.ResolveTCPAddr(t.networkType, t.address)
	if err != nil {
		log.Fatal(err)
	}
	listener, err := net.ListenTCP(t.networkType, tcpAddr)
	if err != nil {
		log.Fatal(err)
	}
	//开启 tcp 服务
	go t.srv.Start(listener, time.Second)
	log.Println("tcp  listening:", listener.Addr())
}

func (t *TcpServer) GetRawSrv() *gotcp.Server {
	return t.srv
}

//----------------------数据流的处理逻辑------------------------
type tcpHandleProcess struct {
	handle tcpHandle
}

//socket建立连接 处理
func (t *tcpHandleProcess) OnConnect(c *gotcp.Conn) bool {

	addr := c.GetRawConn().RemoteAddr()
	c.PutExtraData(addr)
	log.Println("OnConnect:", addr.String())

	//建立档案
	a := archives.NewconnArchive("tcp", c)
	archives.ConnArchStore[addr.String()] = a
	return true
}

//socket关闭连接 处理
func (t *tcpHandleProcess) OnClose(c *gotcp.Conn) {


	log.Println("OnClose:", c.GetExtraData())

	remoteAddr := c.GetRawConn().RemoteAddr().String()

	//因为关闭连接，所以对于网关而言，需要有多个设备下线
	for _,e := range archives.GhostStore{
		if e.Connection != nil && e.Connection.Conn == c {
			e.ClearConnection()
		}
	}

	//删除连接档案
	delete(archives.ConnArchStore, remoteAddr)
}

//上行处理
func (t *tcpHandleProcess) OnMessage(c *gotcp.Conn, p gotcp.Packet) bool {

	// 1. 获取连接的本地端口号
	locAddr := c.GetRawConn().LocalAddr().String()
	port := strings.Split(locAddr, ":")[1]
	remoteAddr := c.GetRawConn().RemoteAddr().String()

	// 2. 搜索端口对应的应用
	var app *archives.AppArchive = nil
	for _, a := range archives.AppArchStore {
		if port == a.ConnPort {
			app = a
			break
		}
	}

	// 3. 根据应用类型查找影子设备
	if app != nil {

		switch app.AppType {
		case "gateway":
			{
				id := t.handle.GetEquipId(p)	//获取设备报文中的id
				if id != ""{
					for i:=0;i<len(app.EquipList);i++{
						if app.EquipList[i].Feature == id{	//找到设备
							equipName := app.EquipList[i].EquipName
							ghostEquip := archives.GhostStore[equipName] //找到影子设备
							if ghostEquip.Connection == nil { 		//若影子设备下，没有将设备档案与连接档案绑定，则此处进行绑定
								ghostEquip.SetConnection(archives.ConnArchStore[remoteAddr])
							}
							t.handle.UplinkHandle(p)
							break
						}
					}
				} else{
					log.Println("gateway need equip id!")
				}

			}
		case "iot":
			{
				// 查找「第一个」设备的影子设备
				if len(app.EquipList) > 0 {
					equipName := app.EquipList[0].EquipName	//应用下第一个设备名称
					ghostEquip := archives.GhostStore[equipName] //找到影子设备
					if ghostEquip.Connection == nil { 		//若影子设备下，没有将设备档案与连接档案绑定，则此处进行绑定
						ghostEquip.SetConnection(archives.ConnArchStore[remoteAddr])
					}
					t.handle.UplinkHandle(p)
				}

			}
		}
	}



	return true
}

//上行解包
func (t *tcpHandleProcess) ReadPacket(conn *net.TCPConn) (gotcp.Packet, error) {

	for {
		data := make([]byte, 1024)

		readLengh, err := conn.Read(data)

		if err != nil { //EOF, or worse
			return nil, err
		}

		if readLengh == 0 { // Connection maybe closed by the client
			return nil, gotcp.ErrConnClosing
		} else {

			//解包程序
			result, err := t.handle.UplinkParse(data[:readLengh])

			return result, err
		}
	}
}

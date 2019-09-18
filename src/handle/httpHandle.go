package handle

import (
	"IoTServer/src/archives"
	"IoTServer/src/gotcp"
	"IoTServer/src/protocol"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)


func SearchEquip (w http.ResponseWriter, r *http.Request, _ httprouter.Params){

	//搜索设备，对所有的连接执行所有应用的搜设备命令
	for _,c := range archives.ConnArchStore{
		for _, a := range archives.AppArchStore{


			var p []gotcp.Packet
			switch a.Protocol {
			case "ElectricMeter":
				p = protocol.SearchEMEquipCommand()
			case "Temperature":
				p = protocol.SearchTempEquipCommand()
			}
			for i:=0;i<len(p);i++{
				c.SendMessage(p[i])
			}
		}
	}
	fmt.Fprint(w, "Search Over!\n")
}


func SendMessage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	ip := r.PostFormValue("ip")
	info := r.PostFormValue("info")
	proto := r.PostFormValue("protocol")

	conn := archives.ConnArchStore[ip] //获取档案
	if conn != nil{
		var p gotcp.Packet
		switch proto {
		case "ElectricMeter":
			p = protocol.NewElectricMeterDownPacket(info)	//组合有功总电能
		}
		conn.SendMessage(p)

		fmt.Fprint(w, "Send Success!\n")

	} else{

		fmt.Fprint(w, "Equip can not connect\n")
	}

}

func GhostOperation(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	name := r.PostFormValue("name")
	info := r.PostFormValue("info")	//信息

	ghostEquip := archives.GhostStore[name]	//获取影子设备
	if ghostEquip == nil{
		fmt.Fprint(w, "Equip can not find\n")
	} else{
		proto := archives.AppArchStore[ghostEquip.AppName].Protocol	//获取协议
		conn := ghostEquip.Connection	//获取影子设备存储的连接档案
		if conn == nil{
			fmt.Fprint(w, "Equip is offLine\n")
		} else{
			var p gotcp.Packet
			switch proto {
			case "ElectricMeter":
				p = protocol.NewElectricMeterDownPacket(info)	//组合有功总电能
			case "Temperature":
				p = protocol.NewTemperatureDownPacket(info)
			}
			conn.SendMessage(p)
			fmt.Fprint(w, "Send Success!\n")
		}
	}
}


func Broadtest (w http.ResponseWriter, r *http.Request, _ httprouter.Params){

	info := r.PostFormValue("info")	//信息
	//搜索设备，对所有的连接执行所有应用的搜设备命令
	for _,c := range archives.ConnArchStore{
		for _, a := range archives.AppArchStore{

			var p gotcp.Packet
			switch a.Protocol {
			case "ElectricMeter":
				p = protocol.NewElectricMeterDownPacket(info)	//组合有功总电能
			case "Temperature":
				p = protocol.NewTemperatureDownPacket(info)
			}
			log.Println("send broad: ")
			c.SendMessage(p)
			break
		}
	}
	fmt.Fprint(w, "Search Over!\n")
}


type httpServer struct {
	port string
	router *httprouter.Router
}

func NewHttpServer(p string) *httpServer{
	return &httpServer{
		port: p,
	}
}

func (h *httpServer) Start() {

	h.router = httprouter.New()
	h.router.POST("/SendMessage/", SendMessage)
	h.router.POST("/GhostOperation/", GhostOperation)
	h.router.GET("/SearchEquip/", SearchEquip)


	h.router.POST("/Broadtest/", Broadtest)


	log.Println("http listening:  :", h.port)
	log.Fatal(http.ListenAndServe(":"+h.port, h.router))
}

func (h *httpServer) Close(){
	h.Close()
}
package main

import (
	"IoTServer/src/archives"
	"IoTServer/src/handle"
	"IoTServer/src/protocol"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) //设置并发执行时使用的CPU的数目

	// 连接档案/应用档案缓存初始化
	archives.ConnArchStore = make(map[string]*archives.ConnArchive)
	archives.AppArchStore = make(map[string]*archives.AppArchive)
	archives.GhostStore = make(map[string]*archives.GhostArchive)
	// 读取应用档案文件夹，获取应用数目
	var s []string
	//s, _ = GetAllFile("/Users/jerry/Desktop/IoTServer/bin/xml", s)
	s, _ = GetAllFile("./xml", s)
	// a. 读取应用档案文件
	for i := 0; i < len(s); i++ {
		arch, err := archives.NewAppArchive(s[i])
		if err != nil {
			log.Fatal("read xml err: ", err)
		}
		archives.AppArchStore[arch.AppName] = arch
	}

	// b. 根据应用档案开启服务
	for _, a := range archives.AppArchStore {

		nettype := a.ConnMode
		addr := ":" + a.ConnPort
		switch a.Protocol {
		case "ElectricMeter":
			htcp := handle.NewTcpHandleProcess(nettype, addr, &protocol.ElectricMeterHandle{})
			//a.SetRouteServer(htcp.GetRawSrv())
			go htcp.Start()
		case "Temperature":
			htcp := handle.NewTcpHandleProcess(nettype, addr, &protocol.TemperatureHandle{})
			go htcp.Start()
		default:
			log.Fatal("protocol can not find!")
		}
	}

	// c. 根据应用档案生成影子设备
	for _, a := range archives.AppArchStore {
		for i := 0; i < len(a.EquipList); i++ {
			archives.GhostStore[a.EquipList[i].EquipName] = archives.NewGhostArchive(&a.EquipList[i], a.AppName)
		}
	}

	http := handle.NewHttpServer("8080")
	go http.Start()

	//捕获信号
	chSig := make(chan os.Signal)
	signal.Notify(chSig, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Signal: ", <-chSig)

	//关闭服务
	//for _, a := range archives.AppArchStore {
	//	a.CloseServer()
	//}

}

func GetAllFile(pathname string, s []string) ([]string, error) {
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		fmt.Println("read dir fail:", err)
		return s, err
	}
	for _, fi := range rd {
		if fi.IsDir() {
			fullDir := pathname + "/" + fi.Name()
			s, err = GetAllFile(fullDir, s)
			if err != nil {
				fmt.Println("read dir fail:", err)
				return s, err
			}
		} else {
			fullName := pathname + "/" + fi.Name()
			s = append(s, fullName)
		}
	}
	return s, nil
}

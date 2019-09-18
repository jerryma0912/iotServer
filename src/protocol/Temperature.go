package protocol

import (
	"IoTServer/src/gotcp"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
)




//------------上下行数据包---------------
// ***上行数据解包后数据结构
type temperatureUpPacket struct{
	//EquipNo string `json:"iotEquipName"`
	Result  string `json:"result"`
}

func (p *temperatureUpPacket) Serialize() []byte{ //上行数据在读取的时候被序列化
	b, _ := json.Marshal(*p)
	return b
}

func NewTemperatureUpPacket (result string) *temperatureUpPacket {	//no string,
	return &temperatureUpPacket{
	//	EquipNo: no,
		Result:  result,
	}
}

// ***下行数据封包后数据结构
type temperatureDownPacket struct {
	pData string
}

func (p *temperatureDownPacket) Serialize() []byte {	//下行数据在写入的时候被序列化

	data, err := hex.DecodeString(p.pData)
	if err != nil {
		log.Println("SendPacket: Hex Command Err")
		return nil
	}
	return data
}

func NewTemperatureDownPacket(str string) *temperatureDownPacket{
	return &temperatureDownPacket{
		pData: str,
	}
}

//------------搜索设备命令------------
func SearchTempEquipCommand() []gotcp.Packet{
	p := make([]gotcp.Packet,1)
	p[0] = NewTemperatureDownPacket("0A0300640006856C")
	return p
}
//------------处理方法----------------
type TemperatureHandle struct {
}


func (e *TemperatureHandle) GetEquipId (p gotcp.Packet) string{

	return ""
}

func (e *TemperatureHandle) UplinkParse (input []byte) (gotcp.Packet, error){

	r := hex.EncodeToString(input)
	fmt.Println(r)

	return NewTemperatureUpPacket("test"),nil
}

func (e *TemperatureHandle) UplinkHandle (packet gotcp.Packet){

	var f temperatureUpPacket
	err := json.Unmarshal(packet.Serialize(), &f)
	if err != nil{
		log.Println("err")

	} else{

		fmt.Println(" 当前温度: ",f.Result)
	}

}
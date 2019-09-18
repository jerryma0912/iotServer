package protocol

import (
	"IoTServer/src/gotcp"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
)




//------------上下行数据包---------------
// ***上行数据解包后数据结构
type eletricMeterUpPacket struct{
	EquipNo string `json:"iotEquipName"`
	Result  string `json:"result"`
}

func (p *eletricMeterUpPacket) Serialize() []byte{	//上行数据在读取的时候被序列化
	b, _ := json.Marshal(*p)
	return b
}

func NewElectricMeterUpPacket(no string, result string) *eletricMeterUpPacket{
	return &eletricMeterUpPacket{
		EquipNo: no,
		Result:  result,
	}
}

// ***下行数据封包后数据结构
type electricMeterDownPacket struct {
	pData string
}

func (p *electricMeterDownPacket) Serialize() []byte {	//下行数据在写入的时候被序列化

	data, err := hex.DecodeString(p.pData)
	if err != nil {
		log.Println("SendPacket: Hex Command Err")
		return nil
	}
	return data
}

func NewElectricMeterDownPacket(str string) *electricMeterDownPacket{
	return &electricMeterDownPacket{
		pData: str,
	}
}

//------------搜索设备命令------------
func SearchEMEquipCommand() []gotcp.Packet{
	p := make([]gotcp.Packet,1)
	p[0] = NewElectricMeterDownPacket("fefefefe68aaaaaaaaaaaa68110433333433ae16")
	return p
}
//------------处理方法----------------
type ElectricMeterHandle struct {
}


func (e *ElectricMeterHandle) GetEquipId (p gotcp.Packet) string{
	var f eletricMeterUpPacket
	err := json.Unmarshal(p.Serialize(), &f)
	if err != nil{
		log.Println("err")
		return ""

	} else{

		return f.EquipNo
	}

}

func (e *ElectricMeterHandle) UplinkParse (input []byte) (gotcp.Packet, error){

	r := hex.EncodeToString(input)
	fmt.Println(r)
	//开始解析
	parseKey := byte(0x33)

	startPos := 4 	//解析协议的起始位(第一个68)
	//中间部分省略
	//dataLength := int(input[startPos+9] - 5)	//数据域长度

	var equipNo string
	equipNo += string2CertainLen(strconv.Itoa(byte2string(input[startPos+6])))
	equipNo += string2CertainLen(strconv.Itoa(byte2string(input[startPos+5])))
	equipNo += string2CertainLen(strconv.Itoa(byte2string(input[startPos+4])))
	equipNo += string2CertainLen(strconv.Itoa(byte2string(input[startPos+3])))
	equipNo += string2CertainLen(strconv.Itoa(byte2string(input[startPos+2])))
	equipNo += string2CertainLen(strconv.Itoa(byte2string(input[startPos+1])))

	if  input[startPos+10] == byte(0)+parseKey &&
		input[startPos+11] == byte(0)+parseKey &&
		input[startPos+12] == byte(1)+parseKey &&
		input[startPos+13] == byte(0)+parseKey {	//当前正向有功T00电能

		var result float64
		result += float64(byte2string(input[startPos+14+3]-parseKey))*10000
		result += float64(byte2string(input[startPos+14+2]-parseKey))*100
		result += float64(byte2string(input[startPos+14+1]-parseKey))*1
		result += float64(byte2string(input[startPos+14+0]-parseKey))*0.01

		return NewElectricMeterUpPacket(equipNo,strconv.FormatFloat(result, 'f', 2, 64)), nil

	}

	return nil, errors.New("uplink Parse err")
}

func (e *ElectricMeterHandle) UplinkHandle (packet gotcp.Packet){

	var f eletricMeterUpPacket
	err := json.Unmarshal(packet.Serialize(), &f)
	if err != nil{
		log.Println("err")

	} else{

		fmt.Println("设备：",f.EquipNo," 当前正向有功T00电能: ",f.Result)
	}

}

func byte2string(b byte) int{
	var result int
	result += int(b & 0xF0)/16*10
	result += int(b & 0x0F)
	return result
}

func string2CertainLen(str string) string{

	if len(str) <2 {
		str = "0" + str
	}
	return str
}
package archives

import (
	"IoTServer/src/etree"
	"errors"
)

type Equip struct {
	EquipName string
	Feature   string
	Option    []map[string]string
}

func NewEquip(name string, feature string, option []map[string]string) *Equip {
	return &Equip{
		EquipName: name,
		Feature:   feature,
		Option:    option,
	}
}

type AppArchive struct {
	AppName   string
	AppType   string
	ConnMode  string
	Protocol  string
	ConnPort  string
	EquipList []Equip
	//routeServer interface{}	//本地启动的服务协程
}

//func (a *AppArchive) SetRouteServer(srv interface{}){
//	a.routeServer = srv
//}
//
//func (a *AppArchive) CloseServer(){
//	switch a.ConnMode {
//	case "tcp":
//		a.routeServer.(*gotcp.Server).Stop()
//	}
//}

func NewAppArchive(path string) (*AppArchive, error) {

	doc := etree.NewDocument()
	if err := doc.ReadFromFile(path); err != nil {
		return nil, errors.New("can not open xml file")
	}

	root := doc.SelectElement("application")
	appname, err1 := xmlElement2String(root, "appname")
	apptype, err2 := xmlElement2String(root, "apptype")
	connmode, err3 := xmlElement2String(root, "connmode")
	protocol, err4 := xmlElement2String(root, "protocol")
	connport, err5 := xmlElement2String(root, "connport")
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
		return nil, errors.New("app param error")
	}

	equiproot := root.SelectElements("equip")
	equipStore := make([]Equip, len(equiproot)) //设备存储
	for j, e := range equiproot {

		equipname, err6 := xmlElement2String(e, "equipname") //设备名称
		if err6 != nil {
			return nil, errors.New("equipname error")
		}

		//特征值
		feature, err7 := xmlElement2String(e, "feature") //设备名称
		if err7 != nil {
			return nil, errors.New("feature error")
		}

		//选项
		option := e.SelectElements("option")
		opStore := make([]map[string]string, len(option)) //选项的存储
		for i, o := range option {
			op, err8 := xmlElement2map(o)
			if err8 != nil {
				return nil, errors.New("option error")
			}
			opStore[i] = op
		}

		equipStore[j] = *NewEquip(equipname, feature, opStore)
	}

	return &AppArchive{
		AppName:   appname,
		AppType:   apptype,
		ConnMode:  connmode,
		Protocol:  protocol,
		ConnPort:  connport,
		EquipList: equipStore,
	}, nil
}

func xmlElement2String(root *etree.Element, element string) (string, error) {

	var s string
	if e := root.SelectElement(element); e != nil {
		s = e.Text()
	} else {
		return "", errors.New("parse xml param err")
	}

	return s, nil
}

func xmlElement2map(root *etree.Element) (map[string]string, error) {

	str1 := root.SelectElement("key").Text()
	str2 := root.SelectElement("value").Text()
	if str1 == "" || str2 == "" {
		return nil, errors.New("parse xml map err")
	}
	m := map[string]string{str1: str2}
	return m, nil
}

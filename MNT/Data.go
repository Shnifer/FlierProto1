package mnt

import (
	"github.com/Shnifer/flierproto1/v2"
	"log"
	"encoding/json"
	"strconv"
)

//Пока обстрактная масса звезда-планета
//TODO: рассмотреть на её примере позже схему композиции
type Star struct{
	//Положение и скорость в абсолютных космических координатах
	Pos V2.V2

	//Скорость. Абсолютных единиц в секунду
	Dir V2.V2

	//Радиус "коллизии" в обсолютных координатах
	ColRad float32

	//Масса, физические взаимодействия
	Mass float32

	//Некие ценные данные её изучения, имитируем цветом и текстом
//	ObservСolor sdl.Color
	ObservText string
}

//TODO: пока глобальным объектом
const GalaxyDataNetPart = 100
var GalaxyData []*Star

//ЗАПУСКАЕТСЯ клиентом, желающим залутать актуальную карту
func DownloadGalaxy() {
	res,err:=Client.CommandResult(CMD_GETGALAXY)
	if err!=nil{
		log.Panicln(err)
	}
	NParts,err := strconv.Atoi(res)
	if err!=nil{
		log.Panicln("DownloadGalaxy, num of Parts ",res,err)
	}
	for i:=0; i<NParts; i++{
		size,err:=strconv.Atoi(<-Client.Recv)
		if err!=nil{
			log.Panicln("DownloadGalaxy, Partsize ",res,err)
		}
		part:=<-Client.Recv

		partData:=make([]*Star, size)
		json.Unmarshal([]byte(part), &partData)
		GalaxyData = append(GalaxyData,partData...)
	}

	log.Println("downloaded galaxy size", len(GalaxyData))
	//for _,v:=range GalaxyData{
	//	log.Println(v)
	//}

}

//Запускается сервером, возвращает строку, которая должна уйти результатом клиенту
func UploadGalaxy() []string{

	L:=len(GalaxyData)
	if (L==0) {
		return []string{"0"}
	}
	//Отправляет по частям,
	NParts := ((L-1) / GalaxyDataNetPart)+1

	res:=make([]string,1+NParts*2) //+1 на сообщении о количестве
	res[0]=strconv.Itoa(NParts)

	for i:=0; i<NParts; i++ {
		size :=  GalaxyDataNetPart
		if i==NParts-1 {
			size = L-GalaxyDataNetPart*(NParts-1)
		}
		startind:=i*GalaxyDataNetPart
		log.Println("part",i,"start",startind,"size",size)
		buf, err := json.Marshal(GalaxyData[startind:startind+size])
		if err != nil {
			log.Panicln(err)
		}
		res[1+2*i] = strconv.Itoa(size)
		res[1+2*i+1] = string(buf)
	}
	return res
}
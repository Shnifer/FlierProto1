package MNT

import (
	"github.com/Shnifer/FlierProto1/V2"
	"log"
	"encoding/json"
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
var GalaxyData []*Star

//ЗАПУСКАЕТСЯ клиентом, желающим залутать актуальную карту
func DownloadGalaxy() {
	res,err:=Client.CommandResult(CMD_GETGALAXY)
	if err!=nil{
		log.Panicln(err)
	}
	json.Unmarshal([]byte(res), &GalaxyData)
}

//Запускается сервером, возвращает строку, которая должна уйти результатом клиенту
func UploadGalaxy() string{
	buf, err:= json.Marshal(GalaxyData)
	if err!=nil{
		log.Panicln(err)
	}
	return string(buf)
}
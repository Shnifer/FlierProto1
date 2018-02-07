package mnt

import (
	"encoding/json"
	"github.com/Shnifer/flierproto1/v2"
	"log"
	"strconv"
)

//Сугубо для целей простоты обмена и маршалинга,
//вне сетевых взаимодействий не использовать
type Color struct {
	R, G, B byte
}

//Пока обстрактная масса звезда-планета
//TODO: рассмотреть на её примере позже схему композиции
type Star struct {
	//Для взаимосвязей планет
	//Можно не указывать, тогда на звезду сложно сослаться
	//Для именных должен быть уникальным
	ID     string
	Parent string

	//Положение и скорость в абсолютных космических координатах
	//для НЕ ПРИВЯЗАННЫХ объектов Parent=""
	Pos V2.V2
	Dir V2.V2
	//Дистанция в абс. коорд. и скорость вращения град/сек вокруг родителя для спутников
	Angle    float32
	OrbDist  float32
	OrbSpeed float32

	//Радиус "коллизии" в обсолютных координатах
	ColRad float32

	//Масса, физические взаимодействия
	Mass float32

	//Некие ценные данные её изучения, имитируем цветом и текстом
	Color      Color
	ObservText string
}

//TODO: пока глобальным объектом
const GalaxyDataNetPart = 100

var GalaxyData []*Star

//ЗАПУСКАЕТСЯ клиентом, желающим залутать актуальную карту
func DownloadGalaxy() {
	res, err := Client.CommandResult(CMD_GETGALAXY)
	if err != nil {
		log.Panicln(err)
	}
	NParts, err := strconv.Atoi(res)
	if err != nil {
		log.Panicln("DownloadGalaxy, num of Parts ", res, err)
	}
	for i := 0; i < NParts; i++ {
		size, err := strconv.Atoi(<-Client.Recv)
		if err != nil {
			log.Panicln("DownloadGalaxy, Partsize ", res, err)
		}
		part := <-Client.Recv

		partData := make([]*Star, size)
		json.Unmarshal([]byte(part), &partData)
		GalaxyData = append(GalaxyData, partData...)
	}

	log.Println("downloaded galaxy size", len(GalaxyData))
	//for _,v:=range GalaxyData{
	//	log.Println(v)
	//}

}

//Запускается сервером, возвращает строки, которая должны уйти результатом клиенту
func UploadGalaxy() []string {

	L := len(GalaxyData)
	if L == 0 {
		return []string{"0"}
	}
	//Отправляет по частям,
	NParts := ((L - 1) / GalaxyDataNetPart) + 1

	res := make([]string, 1+NParts*2) //+1 на сообщении о количестве
	res[0] = strconv.Itoa(NParts)

	for i := 0; i < NParts; i++ {
		size := GalaxyDataNetPart
		if i == NParts-1 {
			size = L - GalaxyDataNetPart*(NParts-1)
		}
		startind := i * GalaxyDataNetPart
		log.Println("part", i, "start", startind, "size", size)
		buf, err := json.Marshal(GalaxyData[startind : startind+size])
		if err != nil {
			log.Panicln(err)
		}
		res[1+2*i] = strconv.Itoa(size)
		res[1+2*i+1] = string(buf)
	}
	return res
}

//Часть main.go ответсвенная за фпс
package fps

import (
	"log"
	"time"
)

type FpsData struct {
	Graph, Phys, Io     int
	MaxDt               float32
	MaxGraphT, MaxPhysT float32
}

type ControlTickerData struct {
	NewGraphPeriodms, NewPhysPeriodms float32
}

type InitStruct struct{
	MIN_FRAME_MS                  int
	MIN_PHYS_MS                   int
	MAX_FRAME_MS                  int
	MAX_PHYS_MS                   int
	TickerBalancerOverhead        float32
	FPS_UPDATE_MS				int
}

var params InitStruct
//Внутренний встроенный тикер, регулярно пересоздаётся на новых частотах
var graphTick,physTick *time.Ticker
//Внешние, постоянные и буферизованные каналы
var GTick, PTick chan time.Time

func Start(initdata InitStruct) (
	ShowfpsTick <-chan time.Time,
	fpsControl chan<- FpsData,
	){

	params = initdata

	ShowfpsTick = time.Tick(time.Duration(params.FPS_UPDATE_MS) * time.Millisecond)

	fpsControl = ListenAndShowFPS()

	GTick=make(chan time.Time,1)
	PTick=make(chan time.Time,1)

	ResetTickersAndStartListen(float32(params.MAX_FRAME_MS),
							float32(params.MAX_PHYS_MS))

	return
}

func ResetTickersAndStartListen (GPeriodMs, PPeriodMs float32) {
	//Из миллисекунд в МИКРОсекунды
	graphTick = time.NewTicker(time.Duration(GPeriodMs*1000) * time.Microsecond)
	physTick = time.NewTicker(time.Duration(PPeriodMs*1000) * time.Microsecond)
	startTickerBufListener(graphTick, physTick, GTick, PTick)
}


func ListenAndShowFPS() (chan<- FpsData) {
	inData := make(chan FpsData)

	lastGraph, lastPhys, lastIO := 0, 0, 0

	go func() {
		overhead := params.TickerBalancerOverhead
		for fps := range inData {
			log.Println(
				"Frame/s:", fps.Graph-lastGraph,
				"Phys/s:", fps.Phys-lastPhys,
				"io/s:", fps.Io-lastIO,
				"max dt", fps.MaxDt*1000, "ms",
				"maxGraph:", fps.MaxGraphT*1000, "ms",
				"maxPhys:", fps.MaxPhysT*1000, "ms")
			lastGraph = fps.Graph
			lastPhys = fps.Phys
			lastIO = fps.Io

			//Из секунд внешнего времени в миллисекунды (для совмест. с ини файлом)
			newGraphPeriod := fps.MaxGraphT * overhead * 1000
			PeriodClamp(&newGraphPeriod, params.MIN_FRAME_MS, params.MAX_FRAME_MS)
			newPhysPeriod := fps.MaxPhysT * overhead * 1000
			PeriodClamp(&newPhysPeriod, params.MIN_PHYS_MS, params.MIN_PHYS_MS)

			graphTick.Stop()
			physTick.Stop()
			ResetTickersAndStartListen(newGraphPeriod,newPhysPeriod)
		}
	}()
	return inData
}

func PeriodClamp(period *float32, min,max int) {
	if *period < float32(min) {
		*period = float32(min)
	}
	if *period>float32(max) {
		*period=float32(max)
	}
	if *period < 1 {
		*period = 1
	}
}

//Запускает перекачку из стандартных тикеров в канал с ёмкостью 1
//Т.к. это случается каждую секунду (период пересчёта фпс),
//тикеры пересоздаются, а этаа функция запускается заново, то заканчиваем
// работу каждой конкретной горутины по тайемру
func startTickerBufListener(GraphTick, PhysTick *time.Ticker, GTickBuf, PTickBuf	chan time.Time) {
	go tickerBufListener(GraphTick, GTickBuf)
	go tickerBufListener(PhysTick, PTickBuf)
}

func tickerBufListener(ticker *time.Ticker, tickBuf chan time.Time){
	stop:=time.After(time.Duration(params.FPS_UPDATE_MS)*time.Millisecond+
		100*time.Millisecond) //живёт чуть дольше, чтобы дослушать
	for{
		select{
		case v:=<-ticker.C:
			tickBuf<-v
		case <-stop:
			return
		}
	}
}
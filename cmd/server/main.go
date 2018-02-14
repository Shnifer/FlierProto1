package main

import (
	"bufio"
	MNT "github.com/Shnifer/flierproto1/mnt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

var NextID func() int

func CreateGenerator() func() int {
	i := 0
	return func() int {
		i++
		return i
	}
}

type inString struct {
	sender net.Conn
	text   string
}

//Комната символизирует корабль,
//Карта соединений по ролям
//Сами комнаты в карте по ИД

type Profile struct {
	Room       string
	Role       string
	rdyForChat bool
}

type Room struct {
	members map[string]net.Conn
}

func newRoom() Room {
	return Room{members: make(map[string]net.Conn)}
}

var BSPForRoom map[string]*MNT.BaseShipParameters

//ГО СИНГЛ сидит на listener.Accept()
func ConnAcepter(listener net.Listener, NewConns chan net.Conn) {
	log.Println("ConnAcepter START")
	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		NewConns <- conn
	}
	log.Println("ConnAcepter END")
}

//ГО 1 на ОТКРЫТОЕ СОЕДИНЕНИЕ сидит на reader.ReadString
func ListenConn(conn net.Conn, inStrs chan inString, DeadConns chan net.Conn) {
	id := NextID()
	log.Println("ListenConn #", id, "START for", conn)
	scaner := bufio.NewScanner(conn)
	for scaner.Scan() {
		str := scaner.Text()
		inStrs <- inString{sender: conn, text: str}
	}
	DeadConns <- conn
	log.Println("ListenConn #", id, "END")
}

//ЗА7ПУСКАЕТ ГОРУТИНУ 1 на СОЕДИНЕНИЕ
//и возвращает канал, из которого принимает строки на отправку
//при закрытии канала выключается
func newSenderConn(conn net.Conn) chan string {
	log.Println("sender for", conn, "STARTED")
	c := make(chan string, 1)

	go func() {
		writer := bufio.NewWriter(conn)
		for msg := range c {
			_, err := writer.WriteString(msg + "\n")
			if err != nil {
				log.Println("SEND to", conn, "ERR:", err)
				continue
			}
			writer.Flush()
			showmsg := msg
			if len(msg) > 40 {
				showmsg = msg[0:40] + "..."
			}
			log.Println("sent", showmsg, "to", conn)
		}
	}()
	c <- MNT.RES_LOGIN
	return c
}

var OutSender map[net.Conn]chan string
var Profiles map[net.Conn]Profile

const serverDataPath = "res/server/"

//Главный тред
func main() {

	LoadDefVals(serverDataPath)

	rand.Seed(time.Now().Unix())

	if DEFVAL.LoadGalaxyFile != "" {
		LoadGalaxyFromFile()
	} else {
		GenerateRandomGalaxy()
	}

	BSPForRoom = make(map[string]*MNT.BaseShipParameters)

	NextID = CreateGenerator()

	NewConns := make(chan net.Conn, 1)
	DeadConns := make(chan net.Conn, 1)
	//	Conns := make(map[net.Conn]bool)

	inStrs := make(chan inString, 1)

	Rooms := make(map[string]Room)
	Profiles = make(map[net.Conn]Profile)

	OutSender = make(map[net.Conn]chan string)

	listener, err := net.Listen("tcp", DEFVAL.tcpPort)
	if err != nil {
		log.Println(err)
	}
	defer listener.Close()

	go ConnAcepter(listener, NewConns)

	for {
		select {
		//Здесь для защиты доступа к Map Conns
		case conn := <-NewConns:
			//START LISTENER
			//Conns[conn] = true
			go ListenConn(conn, inStrs, DeadConns)
		case conn := <-DeadConns:
			//CLOSE & REMOVE FROM CONNS
			conn.Close()
			//delete(Conns, conn)
			prof, ok := Profiles[conn]
			if ok {
				//CHECK OUT from ROOM
				role := prof.Role
				room := prof.Room
				delete(Rooms[room].members, role)
				if len(Rooms[room].members) == 0 {
					//no empty rooms
					delete(Rooms, room)
					delete(BSPForRoom, room)
				}
				//Закрываем канал рассылки и этим горутину
				close(OutSender[conn])
				delete(OutSender, conn)

				log.Println("Profile deleted for mr.", role, "from", room)
				delete(Profiles, conn)
			}

		case inMsg := <-inStrs:
			log.Println("inStrs", inMsg, "come")
			sender := inMsg.sender
			profile, ok := Profiles[sender]
			//Ещё не зарегистрирован, т.е. это должна быть строка логина
			if !ok {
				params := strings.SplitN(inMsg.text, " ", 3)
				if len(params) < 2 {
					//Хуйская строка пропускаем
					log.Println("Conn", sender, " incorrect login:", inMsg.text)
					DeadConns <- sender
					continue
				}

				reqRoom := params[0]
				reqRole := params[1]

				if room, ok := Rooms[reqRoom]; ok {
					if occ, ok := room.members[reqRole]; ok {
						//В этой комнате уже есть такая роль, считаем роли уникальными
						log.Println("Conn", sender, " FAILED login as", reqRole, "in", reqRoom, ". ALREADY OCCUPIED by", occ)
						DeadConns <- sender
						continue
					}
				}

				log.Println("Conn", sender, " SUCCESS login as", reqRole, "in", reqRoom)

				//CHECK IN
				Profiles[inMsg.sender] = Profile{Room: params[0], Role: params[1]}
				room, ok := Rooms[reqRoom]
				if !ok {
					//Если новая комната - добаляем
					room = newRoom()
					Rooms[reqRoom] = room
					BSPForRoom[reqRoom] = LoadBSPFromFile(reqRoom)
				}
				room.members[reqRole] = sender

				//Запускаем отправщика
				ch := newSenderConn(sender)
				OutSender[sender] = ch
			} else {
				//Сообщение от зарегистрированного разбираем
				room := profile.Room
				role := profile.Role

				log.Println("in", room, "mr.", role, "says:", inMsg.text)

				params := strings.SplitN(inMsg.text, " ", 2)
				if len(params) < 1 {
					continue
				}
				command := params[0]
				param := ""
				if len(params) > 1 {
					param = params[1]
				}
				log.Println("command ", command)

				HandleCommand(Rooms[room], sender, profile, command, param, OutSender[sender])
			}
		}
	}
}

//Комната, соединение и роль отправителя уже разобраны. входящая Команда, параметры и канал ответа оправителю
func HandleCommand(Room Room, sender net.Conn, profile Profile, command, params string, out chan string) {
	switch command {
	case MNT.CMD_BROADCAST:
		if len(params) < 2 {
			break
		}
		for destRole, destConn := range Room.members {
			msg := params
			//не шлём себе
			if destRole != profile.Role {
				//Не шлём тем, кто ещё в командном режиме и не готов слушать
				if Profiles[destConn].rdyForChat {
					OutSender[destConn] <- MNT.IN_MSG + " " + msg
				}
			}
		}
	case MNT.CMD_CHECKROOM:
		out <- MNT.RES_CHECKROOM + " " + strconv.Itoa(len(Room.members))

	case MNT.CMD_GETGALAXY:
		res := MNT.UploadGalaxy()
		for _, s := range res {
			out <- s
		}
	case MNT.CMD_READYFORCHAT:
		newprof := Profiles[sender]
		newprof.rdyForChat = true
		Profiles[sender] = newprof
	case MNT.CMD_STOPCHAT:
		newprof := Profiles[sender]
		newprof.rdyForChat = false
		Profiles[sender] = newprof
	case MNT.CMD_GETBSP:
		out <- MNT.RES_BSP + " " + string(BSPForRoom[profile.Room].Encode())
	default:
		log.Println("UNKNOWN COMMAND")
		out <- MNT.ERR_UNKNOWNCMD
	}
}

func LoadBSPFromFile(room string) *MNT.BaseShipParameters {
	//TODO: Сейчас для всех караблей читаем один и тот же файл, должны получать с инженерки или из БД
	var samplData MNT.BaseShipParameters
	exBuf := samplData.Encode()
	if err := ioutil.WriteFile(serverDataPath+"example_"+DEFVAL.LoadBSPFile, exBuf, 0); err != nil {
		log.Panicln("LoadBSPFromFile: can't write even example for you! ", err)
	}
	bsp := new(MNT.BaseShipParameters)
	buf, err := ioutil.ReadFile(serverDataPath + DEFVAL.LoadBSPFile)
	if err != nil {
		log.Panicln("LoadBSPFromFile: can't read BSP for you! ", err)
	}
	bsp.Decode(buf)
	return bsp
}

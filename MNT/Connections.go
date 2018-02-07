package mnt

import (
	"bufio"
	"errors"
	"log"
	"net"
	"time"
)

//TODO: пока один корабль и защитое название
const RoomName = "Firefly"

type ClientConn struct {
	Conn net.Conn
	Send chan string
	Recv chan string
}

const resultTimeOut = time.Second

//Посылает команду и возвращает строку из входного канала,
//ожидая ответ считаем отсутствие его в течении таймаута - ошибкой
func (C ClientConn) CommandResult(com string) (string, error) {
	C.Send <- com
	select {
	case msg := <-C.Recv:
		return msg, nil
	case <-time.After(resultTimeOut):
		return "", errors.New("TIMEOUT for command " + com)
	}
}

func ConnListener(conn net.Conn, Ch chan string) {
	defer close(Ch)
	log.Println("Scanner enabled on server connection")
	scaner := bufio.NewScanner(conn)
	for scaner.Scan() {
		str := scaner.Text()
		log.Println("scaned", str)
		Ch <- str
	}
	if err := scaner.Err(); err != nil {
		log.Println("Scaner error:", err)
	}
	log.Println("Scanner CLOSED on server connection")
}

func ConnSender(conn net.Conn, Ch chan string) {
	log.Println("Writer enabled on server connection")
	writer := bufio.NewWriter(conn)
	for msg := range Ch {
		log.Println("sending", msg)
		if _, err := writer.WriteString(msg + "\n"); err != nil {
			log.Println("Writer CANT SEND to server!", err)
			break
		}
		if err := writer.Flush(); err != nil {
			log.Println("Writer CANT SEND to server!", err)
			break
		}
	}
	log.Println("Writer CLOSED on server connection")
}

func newClientConn(conn net.Conn) ClientConn {
	outCh := make(chan string, 128)
	inCh := make(chan string, 128)

	go ConnListener(conn, inCh)
	go ConnSender(conn, outCh)

	return ClientConn{
		Conn: conn,
		Send: outCh,
		Recv: inCh,
	}
}

//Глобальная перем
var Client ClientConn

func LoginToServer(room, role string) error {
	if Client.Conn == nil {
		return errors.New("No connection established!")
	}

	res, err := Client.CommandResult(room + " " + role)
	if err != nil {
		return err
	}
	if res != RES_LOGIN {
		return errors.New("Login failed! " + res)
	}
	return nil
}

//Соединение - глобальная переменная MNT.Client
func ConnectClientToServer(ServerName, TcpPort string) error {
	const maxtry = 3

	//УСТАНАВЛИВАЕМ СОЕДИНЕНИЕ
	trys := 0
	var conn net.Conn
	for {
		trys++
		c, err := net.Dial("tcp", ServerName+TcpPort)

		if err != nil {
			//При ошибке делаем паузу и пробаем несколько раз
			time.Sleep(200 * time.Millisecond)
			if trys == maxtry {
				return err
			}
		} else {
			conn = c
			break
		}
	}
	log.Println("Connection ", conn)
	Client = newClientConn(conn)
	return nil
}

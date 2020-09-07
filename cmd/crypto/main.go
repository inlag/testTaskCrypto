package main

import (
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/inlag/testTaskCrypto/internal/config"
	"github.com/inlag/testTaskCrypto/internal/database"
	"github.com/inlag/testTaskCrypto/internal/http"
	"github.com/inlag/testTaskCrypto/internal/websocket"
)

func main() {
	var (
		ctrlC = make(chan os.Signal, 1)

		closeRealMain = make(chan struct{})
		errCh         = make(chan error)
	)

	signal.Notify(ctrlC, os.Interrupt)

	go realMain(closeRealMain, errCh)

	for {
		select {
		case err := <-errCh:
			log.Println(err)
		case <-ctrlC:
			closeRealMain <- struct{}{}
			<-closeRealMain
			os.Exit(1)
		}
	}
}

func realMain(ctx chan struct{}, errChan chan error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	dbPool, errInitDB := database.InitDb(config.GetDBUrl())
	if errInitDB != nil {
		log.Fatalln(errInitDB)
	}

	sql := database.Sql{}
	if errSetPool := sql.SetPool(dbPool); errSetPool != nil {
		log.Fatalln(errSetPool)
	}

	webSocket := websocket.WebSocket{}
	errConnectionAndSub := webSocket.ConnectionAndSubscribe()
	if errConnectionAndSub != nil {
		log.Fatalln(errConnectionAndSub)
	}

	var (
		wssToSqlChan, httpToWss = make(chan float64), make(chan float64)
		httpToSql               = make(chan []database.AveragePrice)

		sqlCloseChan, wssCloseChan, httpCloseChan = make(chan struct{}), make(chan struct{}), make(chan struct{})
	)

	server := &http.Server{}
	errSrvInit := server.Initialization(
		net.JoinHostPort(config.GetHost(), config.GetPort()),
		httpToWss,
		httpToSql,
		errChan,
	)
	if errSrvInit != nil {
		log.Fatalln(errSrvInit)
	}

	go sql.Calculate(wssToSqlChan, httpToSql, errChan, sqlCloseChan)

	go webSocket.ReadMessages(wssToSqlChan, httpToWss, errChan, wssCloseChan)

	server.Start()
	go server.Shutdown(httpCloseChan)

	<-ctx

	log.Println("Закрываем веб сокет")
	wssCloseChan <- struct{}{}
	<-wssCloseChan

	log.Println("Закрываем запись в базу данных")
	sqlCloseChan <- struct{}{}
	<-sqlCloseChan

	log.Println("Останавливаем веб сервер")
	httpCloseChan <- struct{}{}
	<-httpCloseChan

	log.Println("Останавливаем приложение")
	ctx <- struct{}{}
}

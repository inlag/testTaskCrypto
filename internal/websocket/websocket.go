package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"github.com/inlag/testTaskCrypto/internal/database"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type WebSocket struct {
	wss      *websocket.Conn
	readChan chan float64
}

func (w *WebSocket) ConnectionAndSubscribe() error {
	errConnection := w.Connection(w.DefaultHost())
	if errConnection != nil {
		return errors.Wrap(errConnection, "create a connection is failed")
	}

	errSubscribe := w.Subscribe()
	if errSubscribe != nil {
		return errors.Wrap(errSubscribe, "subscribe is failed")
	}

	return nil
}

func (w *WebSocket) Connection(url string, header http.Header) error {
	var errDial error
	w.wss, _, errDial = websocket.DefaultDialer.Dial(url, header)
	if errDial != nil {
		return errDial
	}

	return nil
}

func (w *WebSocket) Subscribe() error {
	if w.wss == nil {
		return errors.New("websocket connection is nil")
	}

	errWrite := w.wss.WriteJSON(map[string]string{
		"action":  "subscribe",
		"channel": "trades",
		"symbol":  "BTC-USD",
	})
	if errWrite != nil {
		return errWrite
	}

	w.readChan = make(chan float64)

	return nil
}

func (w *WebSocket) DefaultHost() (string, http.Header) {
	wssHost := url.URL{
		Scheme: "wss",
		Host:   "ws.prod.blockchain.info",
		Path:   "/mercury-gateway/v1/ws",
	}

	head := http.Header{}
	head.Add("origin", "https://exchange.blockchain.com")

	return wssHost.String(), head
}

func (w WebSocket) Next() ([]byte, error) {
	_, message, errReadMessage := w.wss.ReadMessage()
	return message, errReadMessage
}

func (w *WebSocket) CloseConn() error {
	return w.wss.Close()
}

func (w *WebSocket) ReadMessages(toSqlChan chan float64, fromHttp chan float64, errCh chan error, closeChan chan struct{}) {
	go w.readMessage(errCh)
	log.Println("Запускаем чтение WebSocket")
	var current float64
	for {
		select {
		case <-fromHttp:
			fromHttp <- current
		case <-closeChan:
			errClose := w.CloseConn()
			if errClose != nil {
				errCh <- errors.Wrap(errClose, "close connection is failed")
			}
			closeChan <- struct{}{}
			return
		case current = <-w.readChan:
			toSqlChan <- current
		default:

		}
	}
}

func (w *WebSocket) readMessage(errCh chan error) {
	_, _, _ = w.wss.ReadMessage()
	for {
		var ter = new(database.ReceivedMsg)
		message, errReadMessage := w.Next()
		if errReadMessage != nil {
			errCh <- errors.Wrap(errReadMessage, "reading a message is failed")
			continue
		}

		errUnmarshal := json.Unmarshal(message, ter)
		if errUnmarshal != nil {
			errCh <- errors.Wrap(errUnmarshal, "unmarshal message is failed")
			continue
		}
		w.readChan <- ter.Price
	}
}

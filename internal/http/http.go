package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/inlag/testTaskCrypto/docs"
	"github.com/inlag/testTaskCrypto/internal/database"

	"github.com/pkg/errors"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Swagger testTaskCrypto
// @version 1.0
// @description Мини сервер для получения текущей цены BTC, а так же средней стоимости в диапазоне 5 минут за сутки.

// @contact.name Sergey Khorbin
// @contact.email inlag333@yandex.ru

type Server struct {
	routes *http.ServeMux
	srv    *http.Server

	toWssChan chan float64
	toSqlChan chan []database.AveragePrice

	errChan chan error
}

func (s *Server) Initialization(hostPort string, wssChan chan float64, sqlChan chan []database.AveragePrice, errChan chan error) error {
	if wssChan == nil {
		return errors.New("channel for communication with WebSocket is nil")
	}
	if sqlChan == nil {
		return errors.New("channel for communication with SQL is nil")
	}
	if errChan == nil {
		return errors.New("channel for communication with errors is nil")
	}

	s.toSqlChan = sqlChan
	s.toWssChan = wssChan
	s.errChan = errChan

	s.routes = http.NewServeMux()
	s.setRoutes()

	s.srv = &http.Server{
		Addr:    hostPort,
		Handler: s.routes,
	}
	return nil
}

func (s *Server) Start() {
	go func() {
		log.Println("Запускаем Http сервер")
		errListen := s.srv.ListenAndServe()
		if errListen != nil {
			s.errChan <- errListen
		}
	}()
}

func (s *Server) Shutdown(closeChan chan struct{}) {
	<-closeChan
	errShutdown := s.srv.Shutdown(context.Background())
	if errShutdown != nil {
		s.errChan <- errShutdown
	}
	closeChan <- struct{}{}
}

func (s *Server) setRoutes() {
	// Получаем текущую цену BTC
	s.routes.HandleFunc("/price", s.getPriceHandler)
	// Получаем данные цен BTC за день
	// Данные цен состоят из средней за 5 минут с указанием промежутка времени
	s.routes.HandleFunc("/price/data", s.getPriceDataHandler)
	// swagger api
	s.routes.HandleFunc("/", httpSwagger.WrapHandler)
}

type response struct {
	Time  int64   `json:"time"`
	Value float64 `json:"BTC-USD"`
}

// @Summary Возвращает текущую цену BTC.
// @Description Возвращает текущую цену BTC.
// @ID get-price
// @tags Price
// @Produce  json
// @Success 200 {object} response
// @Failure 500 {string} http.StatusInternalServer
// @Router /price [get]
func (s *Server) getPriceHandler(w http.ResponseWriter, r *http.Request) {

	s.toWssChan <- 0
	res := response{
		Time:  time.Now().Unix(),
		Value: <-s.toWssChan,
	}

	resBytes, errMarshal := json.Marshal(res)
	if errMarshal != nil {
		s.errChan <- errMarshal
	}
	w.Write(resBytes)

}

// @Summary Возвращает данные по средней стоимости BTC.
// @Description Возвращает данные по средней стоимости BTC, с интервалом 5 минут.
// @ID get-price-data
// @tags Price
// @Produce  json
// @Success 200 {array} database.AveragePrice
// @Failure 500 {string} http.StatusInternalServer
// @Router /price/data [get]
func (s *Server) getPriceDataHandler(w http.ResponseWriter, r *http.Request) {
	s.toSqlChan <- []database.AveragePrice{}
	res := <-s.toSqlChan
	resBytes, errMarshal := json.Marshal(res)
	if errMarshal != nil {
		s.errChan <- errMarshal
	}
	w.Write(resBytes)

}

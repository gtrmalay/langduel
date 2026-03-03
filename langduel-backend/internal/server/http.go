package server

import (
	"net/http"
)

type Server struct { // Создание сервера
	mux *http.ServeMux //
}

// mux -> диспетчер

func NewRouter() http.Handler {

	s := &Server{ // Создание сервера с диспетчером
		mux: http.NewServeMux(),
	}

	s.routes() // Настройка маршрутов (передача диспетчеру)

	return s.mux // Возвращаем mux, т. к. обрабатывать HTTP умеет именно он
}
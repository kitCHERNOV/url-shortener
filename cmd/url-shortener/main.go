package main

import (
	"fmt"
	"urlsh/internal/config"
)

func main() {
	cfg := config.MustLoad()

	fmt.Println(cfg)

	// TODO: init loger: slog

	// TODO: init storage: sqlite

	// TODO: init router: chi :- полная совместимость с net/http
	//	  chi/render - пакет для рендера
	// TODO: run server:
}

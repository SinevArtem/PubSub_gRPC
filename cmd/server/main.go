package main

import (
	"VK_task/internal/config"
	"VK_task/internal/server"
	"log/slog"
	"os"
)

// здесь мы отдаем приказ приложению запуститься
// точек входа может быть несколько, например из тестов
func main() {
	cfg := config.MustLoad()

	log := setupLogger()

	server.MustRun(*cfg, log)

}

func setupLogger() *slog.Logger {
	log := slog.New(slog.NewTextHandler(
		os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug},
	))

	return log

}

Сервис реализует паттерн "Издатель-Подписчик" (Pub/Sub) через gRPC и состоит из: 

1) subpub (pkg/subpub) - реализует базовую логику подписок и публикаций
2) gRPC сервер (internal/services/pubsub) - предоставляет gRPC интерфейс к ядру
3) Конфигурация (internal/config) - управление параметрами сервера
4) Запуск (cmd/server) - точка входа приложения



Как работает сервис

SubPub:
1) Subscribe(subject string, cb MessageHandler) - подписка на события
2) Publish(subject string, msg interface{}) - публикация события
3) Close(ctx context.Context) - завершение работы сервиса
4) Управляет подписками через map[string][]*subscription
5) Обеспечивает потокобезопасность через sync.Mutex
        
gRPC сервер:
1) Subscribe - позволяет клиенту подписаться на события по определенному ключу и получать их в реальном времени через постоянное соединение.
   Клиент -> SubscribeRequest -> Сервер -> subpub.Subscribe() -> Создается подписка
3) Publish - позволяет клиенту опубликовать событие для всех подписчиков определенного ключа.
   Сервер -> stream.Send() -> Клиент (поток событий)

Сборка:

```
go mod tidy
go build -o bin/pubsub-server cmd/server/main.go
```

Запуск:

```
./bin/pubsub-server --config=./config/local.yaml
```



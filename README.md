# Pena
### Задача
>   Задача: реализовать клиент-серверную архитектуру и покрыть её тестами
>   Сервер:
>   
>   В цикле генерирует структуры типа msg описанную ниже. Сгенерированные структуры отправляет "свободным" клиентам по очереди. Клиент считается свободным, если не получил сообщений или отчитался о выполнении. Отправлять клиенту сообщения в количестве batchsize. В случае, если msg было отправлено, но в report не был получен его id, отправлять повторно, до трёх раз. в случае третьей неудачи, сохранить в bbolt хранилище неуспешное сообщение. Ни один из msg не должен обрабатываться двумя клиентами одновременно
>   имеет 2 эндпоинта: /task и /report
>   /task - принимает через query параметр batchsize и отдаёт клиенту по sse структуры 
>   
>   type msg struct {
>   
>   Id string // https://github.com/rs/xid
>   
>   Period uint64 // 1-1000 random
>   
>   }
>   /report - принимает массив id структур описанных выше и помечает их обарботанными. тем или иным образом, на ваше усмотрение. уже обработанные msg повторно отправляться не должны
>   
>   Клиент:
>   При запуске случайным образом генерирует batchsize от 2 до 10
>   подключается к серверу по sse к эндпоинту /task
>   при получении msg создаёт горутину, которая будет спать Period миллисекунд
>   если пришло msg, но уже запущено batchsize горутин - обработать ошибку и не создавать новую горутину.
>   если Period > 800, горутина должна запаниковать и завершиться, но не клиент.
>   если Period > 900, должен завершиться клиент.
>   
>   как только все batchsize горутин так или иначе завершатся, отправить массив их id на /report
### Структура
```bash
├───client
│   ├───cmd
│   ├───internal
│   │   ├───app
│   │   └───models
│   └───pkg
│       └───sse
└───server
    ├───cmd
    ├───internal
    │   ├───app
    │   ├───handlers
    │   ├───models
    │   └───service
    └───pkg
        └───sse
```

### APIs

/task
/report
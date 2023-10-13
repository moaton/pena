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
│   └───internal
│       ├───app
│       ├───models
│       └───service
└───server
    ├───cmd
    ├───internal
    │   ├───app
    │   ├───db
    │   │   └───bbolt
    │   ├───models
    │   ├───service
    │   └───transport
    │       ├───handlers
    │       └───sse
    └───pkg
        └───util
```

### APIs
------------------------------------------------------------------------------------------

#### Подключение к серверу
<details>
 <summary><code>GET</code> <code><b>/</b></code> <code>(Для подключения к серверу по SSE)</code></summary>
>   После подключения будут отправляться сообщения равные batchsize. На каждое сообщение будет создаваться горутина для проверки Period

##### Ответы

> | http code     | content-type                      | response                                                            |
> |---------------|-----------------------------------|---------------------------------------------------------------------|
> | `200`         | `text/event-stream`               | `{"id":"ckkgmc3h5gbufi2sr320","period":382}`                        |
> | `400`         | `application/json`                | `{"error": ""}`                                                     |

</details>

------------------------------------------------------------------------------------------

#### Отправка репорта

<details>
 <summary><code>POST</code> <code><b>/</b></code> <code>(Для сохранения обработанных сообщений)</code></summary>
>   После того как сервер получит их, он будет удалять из списка и если в списке останутся сообщения они будут переотправлены 3 раза, после 3 раза мы сохраняем в бакет

##### Параметры

> | name      |  type     | data type               | description                                                           |
> |-----------|-----------|-------------------------|-----------------------------------------------------------------------|
> | ids       |  required | object (JSON)           | N/A  |


##### Ответ

> | http code     | content-type                      | response                                                            |
> |---------------|-----------------------------------|---------------------------------------------------------------------|
> | `200`         | `application/json`                | `{message:'success'}`                                               |

</details>

------------------------------------------------------------------------------------------

#### Получение репортов

<details>
 <summary><code>GET</code> <code><b>/</b></code> <code>(Для получения списка обработанных сообщений)</code></summary>


##### Ответ

> | http code     | content-type                      | response                                                            |
> |---------------|-----------------------------------|---------------------------------------------------------------------|
> | `200`         | `application/json`                | `{ids:['ckkg3sm0tshtqjgo9cb0', 'ckkg3qu0tshtqjgo9c7g']}`            |

</details>

------------------------------------------------------------------------------------------

#### Получение неуспешных сообщений

<details>
 <summary><code>GET</code> <code><b>/</b></code> <code>(ДляДля получения неуспешных сообщений)</code></summary>


##### Ответ

> | http code     | content-type                      | response                                                            |
> |---------------|-----------------------------------|---------------------------------------------------------------------|
> | `200`         | `application/json`                | `{ids:['ckkg3sm0tshtqjgo9cb0', 'ckkg3qu0tshtqjgo9c7g']}`            |

</details>

------------------------------------------------------------------------------------------

### Database
>   Как база данных используется bbolt и создается 2 бакета: report и fail, сохраняются обработанные сообщения и не успешные сообщения соответственно
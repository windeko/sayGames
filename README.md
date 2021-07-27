# SayGames

## Тестовое Задание
<i> BY Traigel Vladimir </i>


```
Предлагаю реализовать миниатюрный прототип нашей системы сбора данных и познакомиться с нашими core технологиями.

Клиент посылает аналитические события пачками в виде POST запроса, в теле которого находятся сериализованные в JSON объекты, разделенные \n.
Каждый объект - это событие. Объект имеет такую структуру (в примере отформатированный JSON, в запросе - нет):

{
    "client_time":"2020-12-01 23:59:00",
    "device_id":"0287D9AA-4ADF-4B37-A60F-3E9E645C821E",
    "device_os":"iOS 13.5.1",
    "session":"ybuRi8mAUypxjbxQ",
    "sequence":1,
    "event":"app_start",
    "param_int":0,
    "param_str":"some text"
}

Задача - написать высокопроизводительный сервер для приема таких событий, их обогащения и вставки в БД.
Идеально будет развернуть эту систему на виртуальном сервачке.

Сервер пишем на Go, используя его фичу channels.

В качестве БД используем ClickHouse, который надо поставить, создать в нем базу данных и табличку.

Код сервиса заливаем в любой git.

В процессе обогащения данных, добавляем в событие два поля:

    "ip":"8.8.8.8",
    "server_time":"2020-12-01 23:53:00"

Все, что не указано явно в задании, остается на твое усмотрение.
```

### Запуск
```
docker compose up
```

### Серьезно работает?
Ага. Можно заглянуть в Базу Данных:
```
http://localhost:8123/play

select * from sayGames.logs
```

### Сервисы
- <b> log-generator </b> - Генерирует логи приходящие от клиента пачками. Шлет логи POST-запросом в log-receiver.
- <b> log-receiver </b> - Принимает логи от клиентов, обрабатывает, обогощает и пишет в Clickhouse транзакциями.

### Entrypoints
Скрипты которые используются для отслеживания запуска контейнеров и первичного создания
базы данных и таблицы для хранения логов.

### Что можно было бы улучшить
Я понимаю что это не идеал. Данную реализацию неплохо было бы распилить на модули
и использовать менеджер очередей. Работу с JSON можно так же переложить на easyjson от mailru,
который который работает по принципу кодогенерации и дает отличный прирост к скорости,
но это тестовое задание.
Понять простить.

### Benchmark
```
ab -k -p abjson.txt -T application/json -c 1 -n 100 http://localhost:8080/logs

Server Software:
Server Hostname:        localhost
Server Port:            8080

Document Path:          /logs
Document Length:        0 bytes

Concurrency Level:      1
Time taken for tests:   0.124 seconds
Complete requests:      100
Failed requests:        0
Keep-Alive requests:    100
Total transferred:      9900 bytes
Total body sent:        762100
HTML transferred:       0 bytes
Requests per second:    803.87 [#/sec] (mean)
Time per request:       1.244 [ms] (mean)
Time per request:       1.244 [ms] (mean, across all concurrent requests)
Transfer rate:          77.72 [Kbytes/sec] received
                        5982.72 kb/s sent
                        6060.44 kb/s total
```
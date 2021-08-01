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
Я понимаю что это не идеал. Работу с JSON можно переложить на easyjson от mailru,
который который работает по принципу кодогенерации и дает отличный прирост к скорости,
впилить менеджер очередей, но это тестовое задание.
Понять простить.

### Benchmark
```
ab -k -p abjson.txt -T application/json -c 6 -n 100 -t 5 http://localhost:8080/logs

Server Software:
Server Hostname:        localhost
Server Port:            8080

Document Path:          /logs
Document Length:        0 bytes

Concurrency Level:      6
Time taken for tests:   5.048 seconds
Complete requests:      5490
Failed requests:        0
Keep-Alive requests:    5490
Total transferred:      543510 bytes
Total body sent:        41885016
HTML transferred:       0 bytes
Requests per second:    1087.58 [#/sec] (mean)
Time per request:       5.517 [ms] (mean)
Time per request:       0.919 [ms] (mean, across all concurrent requests)
Transfer rate:          105.15 [Kbytes/sec] received
                        8103.02 kb/s sent
                        8208.17 kb/s total
```
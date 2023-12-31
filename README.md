## SYSTEM MONITORING SERVER

GRPC-сервер сбора статиски состояния системы за заданный период времени.

## Общее описание 

Сервер для каждого клиента каждые N секунд выдает информацию о системе, усредненную за последние M секунд по GRPC протоколу в непрерывном потоке (grpc streaming). N и M - параметры запроса клиента.

Данные, которые передаются сервером:
1. средняя загрузка системы - load average;   
2. загрузка процессора - %user_mode, %system_mode, %idle;
3. загрузка диска(ов):
    - tps (transfers per second)
    - KB/s (kilobytes (read+write) per second)
4. информация о дисках по каждой файловой системе:
    - использовано мегабайт, % от доступного количества;
    - использовано inode, % от доступного количества.
5. статистика по сетевым соединениям:
    - слушающие TCP & UDP сокеты: command, pid, user, protocol, port;
    - количество TCP соединений, находящихся в разных состояниях (ESTAB, FIN_WAIT, SYN_RCV и пр.)
6. top talkers по сети:
    - по протоколам: protocol (TCP, UDP, ICMP, etc), bytes, % от sum(bytes) за последние M), сортируем по убыванию процента
    - по трафику: source ip:port, destination ip:port, protocol, bytes per second (bps), сортируем по убыванию bps
 
Статистика ("снапшот" системы) представляет собой объекты, описанные в формате Protobuf: api/api.proto.

Информация выдается всем подключенным по GRPC клиентам.

Запросы:
- GetSystemDump(N, M) - разовое получение дампа системы за M секунд (N игнорируется)
- StreamSystemDump(N, M) - получение дампа системы за M секунд каждые N секунд (целевое решение)
 
Параметры конфигурации сервера задаются в файле config.json. Файл передается в командной строке.

# Клиент 

Для проведения (локальных) интеграционнх тестов реализован простой клиент (папка client), который в реальном времени получает и выводит в STDOUT (например) сетевую статистику в виде таблицы.

## Реализация

При работе сервера дамп системы формируется каждую секунду и сохраняется в вытесняющий КЭШ (LRU Cash) по размеру. Размер - максимальное значение M из текущего множества запросов клиентов.
Сетевая статистика записывается в вытесняющий КЭШ по времени (полторы секунды).

Данная реализация позволяет не расходовать ОП, но при большой нагрузки на сервис и на сетевой трафик сервера возможна загрузка СПУ на множественную реаллокацию.

Dockerfile собирает проект со статическими параетрами сборки и переносит сборку и config.json в scratch. Результат сборки - distroless-образ.

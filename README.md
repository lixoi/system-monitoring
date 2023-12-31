## SYSTEM MONITORING SERVICE

GRPC-сервер сбора статиски состояния системы за заданный период времени:
- загрузка процессора
- загрузка диска 
- сетевая статистика

Каждые N секунд клиенту выдается дамп (статистика) за M секунд по GRPC протоколу в непрерывном потоке. N и M - параметры запроса.

Запросы:
- GetSystemDump(N, M) - разовое получение дампа системы за M секунд (N игнорируется)
- StreamSystemDump(N, M) - получение дампа системы за M секунд каждые N секунд
 
Параметры конфигурации сервера задаются в файле config.json. Файл передается в командной строке.

При работе сервера дамп системы формируется каждую секунду и сохраняется в вытесняющий КЭШ по размеру. Размер - максимальное значение M из текущего множества запросов клиентов.
Сетевая статистика записывается в вытесняющий КЭШ по времени (полторы секунды).

Данная реализация позволяет не расходовать ОП, но при большой нагрузки на сервис и на сетевой трафик сервера возможна загрузка СПУ на множественную реаллокацию.

Пайтлдайн проекта переделан для ветки system_stats_daemon. Джоба сборки не может отработать, так нет возможности подготовить Docker-образ для раннера: при сборке необходимы зависимости libpcap-dev (см. Dockerfile проекта) 

Dockerfile собирает проект и переносит сборку и config.json в scratch. Результат сборки - distroless-образ.
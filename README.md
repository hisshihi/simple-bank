# О проекте simple-bank

## Создание миграций

```bash
migrate create -ext sql -dir db/migration -seq init_schema

create - создает новую миграцию
ext - расширение файла
dir - директория с миграциями
seq - название миграции
```

## Применение миграций

```bash
migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/simple-bank?sslmode=disable" -verbose up
```

## Создание sql запросов с помощью sqlc

```bash
sqlc init

создание файла sqlc.yaml
```

## Генерация sql запросов

Создаём папку query и в ней создаём файлы sql запросов.

```bash
sqlc generate
```

В models.go будут сгенерированы структуры данных для создания таблиц в базе данных.

В db.go будет сгенерирован интерфейс для работы с базой данных.

- Интерфейс DBTX - Этот интерфейс определяет основные методы для работы с базой данных:

```go
type DBTX interface {
    ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
    PrepareContext(context.Context, string) (*sql.Stmt, error)
    QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
    QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}
```

- ExecContext - выполнение SQL-запросов без возврата данных (INSERT, UPDATE, DELETE)
- PrepareContext - подготовка SQL-запроса
- QueryContext - выполнение запроса с возвратом множества строк
- QueryRowContext - выполнение запроса с возвратом одной строки

- Функция New - Создает новый экземпляр структуры Queries для работы с базой данных.

```go
func New(db DBTX) *Queries {
    return &Queries{db: db}
}
```

В queries.go будут сгенерированы sql запросы.

- Структура Queries - Основная структура, которая содержит соединение с базой данных.

```go
type Queries struct {
    db DBTX
}
```

- Метод WithTx - Позволяет создать новый экземпляр Queries с транзакцией SQL.

```go
func (q *Queries) WithTx(tx *sql.Tx) *Queries {
    return &Queries{
        db: tx,
    }
}
```

## Пример файла sqlc.yaml

```yaml
version: "2"
sql:
    # Определяем схему для PostgreSQL
    - schema: "./db/migration/" # Директория, где хранятся файлы миграций/схемы базы данных
      queries: "./db/query/" # Директория с SQL-запросами
      engine: "postgresql" # Указываем, что используем PostgreSQL

      # Настройки генерации кода
      gen:
          go: # Настройки для Go (можно заменить на другой язык)
              package: "sqlc" # Имя пакета для сгенерированного кода
              out: "./db/sqlc" # Директория для сгенерированного кода
              emit_json_tags: true # Генерировать JSON-теги для структур
              emit_prepared_queries: false # Использовать prepared statements
              emit_interface: true # Генерировать интерфейсы
              emit_exact_table_names: false # Использовать точные имена таблиц
              emit_empty_slices: true # Возвращать пустые слайсы вместо nil

      # Дополнительные настройки для работы с базой данных
      database:
          uri: "postgresql://postgres:postgres@localhost:5432/simple-bank?sslmode=disable"
```

## Тесты

Создаём файл main_test.go в папке sqlc.
Для создания соединения с базой данных и тестовых запросов

Для подключения к базе данных используем пакет

```bash
go get github.com/lib/pq
```

## Тесты для sql запросов

В папке db/sqlc создаём файлы для тестирования sql запросов.
Для большего удобства тестирования запросов, добавляем библиотеку для тестирования.

```bash
go get github.com/stretchr/testify
```

## Mock тесты

Для mock тестов используем библиотеку

```bash
go install go.uber.org/mock/mockgen@latest
```

Для создания mock тестов используем команду

```bash
mockgen -package mockdb -destination db/mock/store.go github.com/hisshihi/simple-bank/db/sqlc Store
```

- package - имя пакета для создания mock тестов
- destination - директория для создания mock тестов
- github.com/hisshihi/simple-bank/db/sqlc - путь к пакету
- Store - интерфейс, который мы хотим протестировать

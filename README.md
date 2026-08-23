
# Pg cpol (PostgreSQL Connection Pool)

Это демонстрациооная версия самонаписанного клиентского ядра взаимодействия с срвером Postgres для Go приложений

Мотивацией послужила библиотека pgx, в следствии чего модуль с маппингом протокола остался почти идентичным, а также, реализация оставляет возможность использовать библиотеку как и one-time-connection в обход использования пула
## Installation

Get pg cpol

```bash
  go get github.com/TiJon8/cpol
```
    
## Usage

```go
  config := map[string]string{
		"host": "localhost",
		"port": "5432",
		"user": "root",
		"password": "root",
		"database": "records",
		"connect_timeout": "15", // seconds
		"pre_open_connections": "1",
		"pool_max_connections": "5",
	}
	poolConfig, err := pgpool.ParseConfigMap(config)
	if err != nil {
		panic(err)
	}
	pool, err := pgpool.InitWithConfig(context.Background(), poolConfig)
	if err != nil {
		panic(err)
	}
	pool.Ping(context.Background()) // check if connection is alive

    ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*999)
	defer cancel()
	sql := `
		SELECT id, name FROM records WHERE name=$1;
	`
	rows, err := pool.Query(ctx, sql, "Name")
	if err != nil {
		fmt.Println(err)
	}
	var id string
	var name string
	for rows.Next() {
		if err := rows.Scan(&id, &name); err != nil {
			fmt.Println(err)
			break
		}
	}
	pool.Close()
```

## API Config

| Parameter | Description                |
| :-------- | :------------------------- |
| `host`    | **Required**. Host name    |
| `port`    | **Required**. Port number  |
| `user`    | **Required**. User name    |
| `password`| **Required**. Password     |
| `database`| **Required**. Database name|
| `connect_timeout`    |  Connect Timeout *in seconds*  |
| `pool_max_connections`    |  Max count of connections (default: 4)   |
| `pre_open_connections`    |  Count of pre-opened connections (default: 0)  |
| `pool_max_conn_idle_timeout`    |  Max idle timeout for connection (default: 2 minutes)   |

## Feedback

Если вы наткнулись на данный репозиторий и у вас есть вдохновение и мотивация его улучшить, прошу, эксперементируйте :)

Буду рад предложениям и идеям в [telegram](https://t.me/stesnyashka)


## Related

[Original PGX Library](https://github.com/jackc/pgx#pgx---postgresql-driver-and-toolkit)


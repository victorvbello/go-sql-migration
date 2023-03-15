## DB
Only for MySQL driver, using connection string like `user01:password01@tcp(localhost:3306)/mydatabase?parseTime=true`

## Commands
**Create**
> Run this to create new migration template file
```sh
go run . create -name "create table user" -dir "./migrations"
```
**Up**
> Run this to execute action up of all migration files
```sh
go run . up -conn "user:pass@tcp(127.0.0.1:3306)/test?parseTime=true"
```
**UpTo**
> Run this to execute action up of specific version
```sh
go run . up-to -conn "user:pass@tcp(127.0.0.1:3306)/test?parseTime=true" -version "20230314185040" 
```
**Down**
> Run this to execute action down of all migration files
```sh
go run . down -conn "user:pass@tcp(127.0.0.1:3306)/test?parseTime=true"  
```
**DownTo**
> Run this to execute action down of all migration files with the version major
```sh
go run . down-to -conn "user:pass@tcp(127.0.0.1:3306)/test?parseTime=true" -version "20230314193200" 
```

## Using this guide
https://www.calhoun.io/database-migrations-in-go/

## Using code as example
https://github.com/joncalhoun/migrate

https://github.com/pressly/goose
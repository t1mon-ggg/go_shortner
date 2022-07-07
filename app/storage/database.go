package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	"github.com/t1mon-ggg/go_shortner/app/helpers"
	"github.com/t1mon-ggg/go_shortner/app/models"
)

//SQL queries
const (
	schemaSQL = `
	CREATE TABLE IF NOT EXISTS "ids" (
		"cookie" VARCHAR(32) NOT NULL UNIQUE PRIMARY KEY,
		"key" VARCHAR(64) NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS "urls" (
	  "id" int4 NOT NULL PRIMARY KEY UNIQUE GENERATED ALWAYS AS IDENTITY (
	INCREMENT 1
	MINVALUE  1
	MAXVALUE 2147483647
	START 1
	CACHE 1
	),	
	  "short" varchar(8) NOT NULL UNIQUE,
	  "long" varchar(255) NOT NULL,
	  "cookie" varchar(32) NOT NULL,
	  "deleted" bool NOT NULL DEFAULT false,
	  CONSTRAINT "cookie" FOREIGN KEY ("cookie") REFERENCES "ids" ("cookie") ON DELETE NO ACTION ON UPDATE NO ACTION
	);
	CREATE UNIQUE INDEX IF NOT EXISTS urls_long_idx ON "urls" ("long" text_ops,"cookie" text_ops) WHERE "deleted"=false;
`
	cookieSelectIDs  = `SELECT "cookie", "key" FROM "ids" WHERE "cookie"='%s'`
	cookieSelectURLs = `SELECT "short", "long", "deleted" FROM "urls" WHERE "cookie"='%s'`
	cookieSearch     = `SELECT COUNT("cookie") FROM "ids" WHERE "cookie"='%s'`
	tagSelect        = `SELECT "short", "long", "deleted" FROM "urls" WHERE "short"='%s'`
	urlSelect        = `SELECT "short" FROM "urls" WHERE "long"='%s' AND "cookie"='%s'`
	writeIDs         = `INSERT INTO "ids" ("cookie", "key") VALUES ($1,$2)`
	writeURLs        = `INSERT INTO "urls" ("cookie", "short", "long") VALUES ($1,$2,$3)`
	tagDelete        = `UPDATE "urls" SET "deleted"=true WHERE "cookie"=$1 AND "short"=$2`
)

//Postgres - struct for postgres implementation
type postgres struct {
	conn string  //строка подключения к базе данных
	db   *sql.DB //дескриптор для работы с базой
}

//NewPostgreSQL - создание ссылки на структуру для работы с базой данных
func NewPostgreSQL(s string) (*postgres, error) {
	log.Println("DSN string:", s)
	db := postgres{conn: s}
	err := db.open()
	if err != nil {
		return nil, err
	}
	log.Println("Successfull connection to PostgreSQL")
	err = db.create()
	if err != nil {
		return nil, err
	}
	return &db, nil
}

//open - подключение к базу данных, создание схемы БД
func (s *postgres) open() error {
	var err error
	s.db, err = sql.Open("postgres", s.conn)
	if err != nil {
		return err
	}
	err = s.create()
	if err != nil {
		return err
	}
	return nil
}

//create - создание схемы БД
func (s *postgres) create() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err := s.db.ExecContext(ctx, schemaSQL)
	if err != nil {
		log.Println("?????!")
		return err
	}
	return nil
}

//Ping - проверка состояния соединения с базой данных
func (s *postgres) Ping() error {
	log.Println("Check connection to PostgreSQL")
	ctx := context.Background()
	connection, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	err := s.db.PingContext(connection)
	if err != nil {
		log.Println("Connection to PostgreSQL failed")
		return err
	}
	log.Println("Connection to PostgreSQL confirmed")
	return nil
}

//Close - закрытие дексриптора базы данных
func (s *postgres) Close() error {
	err := s.db.Close()
	if err != nil {
		return err
	}
	return nil
}

//ReadByCookie - чтение из базы данных
func (s *postgres) ReadByCookie(cookie string) (models.ClientData, error) {
	a := models.ClientData{}
	log.Println("Select from IDs")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := fmt.Sprintf(cookieSelectIDs, cookie)
	log.Printf("Executing \"%s\"\n", query)
	var rowCookie, rowKey string
	err := s.db.QueryRowContext(ctx, query).Scan(&rowCookie, &rowKey)
	if err != nil {
		return models.ClientData{}, err
	}
	a.Cookie = rowCookie
	a.Key = rowKey
	a.Short = make([]models.ShortData, 0)
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query = fmt.Sprintf(cookieSelectURLs, cookie)
	log.Printf("Executing \"%s\"\n", query)
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return a, err
	}
	if rows.Err() != nil {
		return a, err
	}
	for rows.Next() {
		var short, long string
		var deleted bool
		err := rows.Scan(&short, &long, &deleted)
		if err != nil {
			return a, err
		}
		a.Short = append(a.Short, models.ShortData{Short: short, Long: long, Deleted: deleted})
	}
	return a, nil
}

//ReadByURL - чтение из базы данных
func (s *postgres) TagByURL(url, cookie string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := fmt.Sprintf(urlSelect, url, cookie)
	var short string
	err := s.db.QueryRowContext(ctx, query).Scan(&short)
	if err != nil {
		log.Println("!!!!!!!!!!!!!!!!", err)
		return "", err
	}
	return short, nil
}

//ReadByTag - чтение из базы данных
func (s *postgres) ReadByTag(tag string) (models.ShortData, error) {
	m := models.ShortData{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := fmt.Sprintf(tagSelect, tag)
	log.Printf("Executing \"%s\"\n", query)
	var short, long string
	var deleted bool
	err := s.db.QueryRowContext(ctx, query).Scan(&short, &long, &deleted)
	if err != nil {
		if helpers.NoRowsError(err) {
			return models.ShortData{}, nil
		}
		return m, err
	}
	m.Short = short
	m.Long = long
	m.Deleted = deleted
	return m, nil
}

//Write - запись в базы данных
func (s *postgres) Write(data models.ClientData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	var count int
	query := fmt.Sprintf(cookieSearch, data.Cookie)
	err := s.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if count == 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		stmt1, err := tx.PrepareContext(ctx, writeIDs)
		if err != nil {
			return err
		}
		defer stmt1.Close()
		_, err = stmt1.ExecContext(ctx, data.Cookie, data.Key)
		if err != nil {
			return err
		}
	}
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	stmt2, err := tx.PrepareContext(ctx, writeURLs)
	if err != nil {
		return err
	}
	defer stmt2.Close()
	for _, value := range data.Short {
		_, err = stmt2.ExecContext(ctx, data.Cookie, value.Short, value.Long)
		if err != nil {
			if helpers.UniqueViolationError(err) {
				newerr := errors.New("not unique url")
				return newerr
			}
			return err
		}
	}
	tx.Commit()
	return nil
}

//Cleaner - delete task worker creator
func (s *postgres) Cleaner(inputCh <-chan models.DelWorker, workers int) {
	fanOutChs := helpers.FanOut(inputCh, workers)
	for _, fanOutCh := range fanOutChs {
		go s.newWorker(fanOutCh)
	}
}

//deleteTag - mark tag as deleted
func (s *postgres) deleteTag(task models.DelWorker) {
	tx, err := s.db.Begin()
	if err != nil {
		log.Println("Error while create transaction:", err)
		return
	}
	defer tx.Rollback()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stmt, err := tx.PrepareContext(ctx, tagDelete)
	if err != nil {
		log.Println("Error while create statement:", err)
		return
	}
	defer stmt.Close()
	for _, tag := range task.Tags {
		_, err = stmt.ExecContext(ctx, task.Cookie, tag)
		if err != nil {
			log.Println("Error while execute query:", err)
			return
		}
	}
	tx.Commit()
}

//newWorker - delete task worker
func (s *postgres) newWorker(input <-chan models.DelWorker) {
	for task := range input {
		s.deleteTag(task)
	}
}

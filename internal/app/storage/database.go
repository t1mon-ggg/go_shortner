package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	_ "github.com/lib/pq"

	"github.com/t1mon-ggg/go_shortner/internal/app/helpers"
)

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
	  "long" varchar(255) NOT NULL UNIQUE,
	  "cookie" varchar(32)NOT NULL,
		CONSTRAINT "cookie" FOREIGN KEY ("cookie") REFERENCES "ids" ("cookie") ON DELETE NO ACTION ON UPDATE NO ACTION
	)
`
	cookieSelectIDs  = `SELECT "cookie", "key" FROM "ids" WHERE "cookie"='%s'`
	cookieSelectURLs = `SELECT "short", "long" FROM "urls" WHERE "cookie"='%s'`
	cookieSearch     = `SELECT COUNT("cookie") FROM "ids" WHERE "cookie"='%s'`
	tagSearch        = `SELECT COUNT("short") FROM "urls" WHERE "short"='%s'`
	tagSelect        = `SELECT "short", "long" FROM "urls" WHERE "short"='%s'`
	urlSelect        = `SELECT "short" FROM "urls" WHERE "long"='%s'`
	writeIDs         = `INSERT INTO "ids" ("cookie", "key") VALUES ('%s','%s')`
	writeURLs        = `INSERT INTO "urls" ("cookie", "short", "long") VALUES ('%s','%s','%s')`
	tagDelete        = ``
	cookieDelete     = ``
)

type Postgresql struct {
	Conn string  //строка подключения к базе данных
	db   *sql.DB //дескриптор для работы с базой
}

//NewPostgreSQL - создание ссыылки на структуру для работы с базой данных
func NewPostgreSQL(conn string) (*Postgresql, error) {
	db := Postgresql{Conn: conn}
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
func (c *Postgresql) open() error {
	var err error
	c.db, err = sql.Open("postgres", c.Conn)
	if err != nil {
		return err
	}
	err = c.create()
	if err != nil {
		return err
	}
	return nil
}

//create - создание схемы БД
func (c *Postgresql) create() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err := c.db.ExecContext(ctx, schemaSQL)
	if err != nil {
		return err
	}
	return nil
}

//Ping - проверка состояния соединения с базой данных
func (database *Postgresql) Ping() error {
	log.Println("Check connection to PostgreSQL")
	ctx := context.Background()
	connection, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	err := database.db.PingContext(connection)
	if err != nil {
		log.Println("Connection to PostgreSQL failed")
		return err
	}
	log.Println("Connection to PostgreSQL confirmed")
	return nil
}

//Close - закрытие дексриптора базы данных
func (database *Postgresql) Close() error {
	err := database.db.Close()
	if err != nil {
		return err
	}
	return nil
}

//ReadByCookie - чтение из базы данных
func (s *Postgresql) ReadByCookie(cookie string) (helpers.Data, error) {
	a := make(map[string]helpers.WebData)
	log.Println("Select from IDs")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := fmt.Sprintf(cookieSelectIDs, cookie)
	log.Printf("Executing \"%s\"\n", query)
	var rowCookie, rowKey string
	err := s.db.QueryRowContext(ctx, query).Scan(&rowCookie, &rowKey)
	if err != nil {
		return nil, err
	}
	a[rowCookie] = helpers.WebData{Key: rowKey, Short: make(map[string]string)}
	entry := a[rowCookie]
	shorts := entry.Short
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query = fmt.Sprintf(cookieSelectURLs, cookie)
	log.Printf("Executing \"%s\"\n", query)
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var short, long string
		err := rows.Scan(&short, &long)
		if err != nil {
			return nil, err
		}
		shorts[short] = long
	}
	entry.Short = shorts
	a[rowCookie] = entry
	log.Println(a)
	return a, nil
}

//ReadByURL - чтение из базы данных
func (s *Postgresql) TagByURL(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := fmt.Sprintf(urlSelect, url)
	var short string
	err := s.db.QueryRowContext(ctx, query).Scan(&short)
	if err != nil {
		return "", err
	}
	return short, nil
}

//ReadByTag - чтение из базы данных
func (s *Postgresql) ReadByTag(tag string) (map[string]string, error) {
	m := make(map[string]string)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := fmt.Sprintf(tagSelect, tag)
	log.Printf("Executing \"%s\"\n", query)
	var short, long string
	err := s.db.QueryRowContext(ctx, query).Scan(&short, &long)
	if err != nil {
		return nil, err
	}
	m[short] = long
	return m, nil
}

//Write - запись в базы данных
func (s *Postgresql) Write(data helpers.Data) error {
	for i := range data {
		ctx1, cancel1 := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel1()
		query := fmt.Sprintf(cookieSearch, i)
		log.Printf("Execuing \"%s\"\n", query)
		var count int
		err := s.db.QueryRowContext(ctx1, query).Scan(&count)
		if err != nil {
			return err
		}
		log.Println("Search cookie result:", count)
		if count == 0 {
			ctx2, cancel2 := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel2()
			query := fmt.Sprintf(writeIDs, i, data[i].Key)
			log.Printf("Execuing \"%s\"\n", query)
			result, err := s.db.ExecContext(ctx2, query)
			if err != nil {
				return err
			}
			affected, err := result.RowsAffected()
			if err != nil {
				return err
			}
			log.Printf("Executed query:\n%s\n has results:\nAffected rows: %d\n", query, affected)
		}
		if len(data[i].Short) != 0 {
			shorts := data[i].Short
			for j := range shorts {
				ctx3, cancel3 := context.WithTimeout(context.Background(), 1*time.Second)
				defer cancel3()
				var counttag int
				query := fmt.Sprintf(tagSearch, j)
				log.Printf("Execuing \"%s\"\n", query)
				err := s.db.QueryRowContext(ctx3, query).Scan(&counttag)
				if err != nil {
					return err
				}
				log.Println("Search cookie result:", counttag)
				if counttag == 0 {
					ctx4, cancel4 := context.WithTimeout(context.Background(), 1*time.Second)
					defer cancel4()
					query := fmt.Sprintf(writeURLs, i, j, shorts[j])
					log.Printf("Execuing \"%s\"\n", query)
					result, err := s.db.ExecContext(ctx4, query)
					if err != nil {
						if driverErr, ok := err.(*pq.Error); ok {
							if pgerrcode.UniqueViolation == driverErr.Code {
								return errors.New("Not UNIQUE URL")
							}
						}
						return err
					}
					affected, err := result.RowsAffected()
					if err != nil {
						return err
					}
					log.Printf("Executed query:\n%s\n has results:\nAffected rows: %d\n", query, affected)
				}
			}
		}
	}
	return nil
}

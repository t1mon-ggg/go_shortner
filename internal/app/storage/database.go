package storage

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/t1mon-ggg/go_shortner/internal/app/helpers"
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
		log.Println("Connection to PostgreSQL failed")
		return nil, err
	}
	log.Println("Successfull connection to PostgreSQL")
	return &db, nil
}

func (database *Postgresql) open() error {
	var err error
	database.db, err = sql.Open("postgres", database.Conn)
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

//Read - чтение из базы данных
func (database *Postgresql) Read() (helpers.Data, error) {
	a := make(map[string]helpers.WebData)
	return a, nil
}

//Write - запись в базы данных
func (database *Postgresql) Write(data helpers.Data) error {
	return nil
}

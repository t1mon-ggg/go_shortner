package storage

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"github.com/t1mon-ggg/go_shortner/internal/app/helpers"
)

type Postgresql struct {
	Conn string  //строка подключения к базе данных
	db   *sql.DB //дескриптор для работы с базой
}

//NewDB - создание ссыылки на структуру для работы с базой данных
func NewDB(conn string) (*Postgresql, error) {
	db := Postgresql{Conn: conn}
	err := db.open(conn)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

func (database *Postgresql) open(connStr string) error {
	var err error
	database.db, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	return nil
}

//Ping - проверка состояния соединения с базой данных
func (database *Postgresql) Ping() error {
	ctx := context.Background()
	connection, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err := database.db.PingContext(connection)
	if err != nil {
		return err
	}
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
	return nil, nil
}

//Write - запись в базы данных
func (database *Postgresql) Write(data helpers.Data) error {
	return nil
}

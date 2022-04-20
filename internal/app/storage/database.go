package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/t1mon-ggg/go_shortner/internal/app/helpers"
)

type Postgresql struct {
	Address string  //адрес сервер базы данных
	db      *sql.DB //дескриптор для работы с базой
}

//NewDB - создание ссыылки на структуру для работы с базой данных
func NewDB(address string) (*Postgresql, error) {
	db := Postgresql{Address: address}
	err := db.open()
	if err != nil {
		return nil, err
	}
	return &db, nil
} //Ping - проверка состояния соединения с базой данных
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

func (database *Postgresql) open() error {
	var err error
	addr := strings.Split(database.Address, ":")
	host := addr[0]
	port := addr[1]
	connStr := fmt.Sprintf("user=postgres password=mypass dbname=productdb sslmode=disable host=%s port=%s", host, port)
	database.db, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	return nil
}

func (database *Postgresql) Close() error {
	var err error
	err = database.db.Close()
	if err != nil {
		return err
	}
	return nil
}

func (database *Postgresql) Read() (helpers.Data, error) {
	return nil, nil
}

func (database *Postgresql) Write(data helpers.Data) error {
	return nil
}

package storage

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/t1mon-ggg/go_shortner/app/models"
)

var testDSN = "postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable"

func dbClean(t *testing.T) {
	s, err := NewPostgreSQL(testDSN)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = s.db.ExecContext(ctx, "TRUNCATE TABLE urls, ids;")
	require.NoError(t, err)
	err = s.Close()
	require.NoError(t, err)
}

func dbPreparation(t *testing.T) {
	dbClean(t)
	s, err := NewPostgreSQL(testDSN)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = s.db.ExecContext(ctx, `INSERT INTO public.ids (cookie,"key") VALUES
	('cookie1','secret-key1'),
	('cookie2','secret-key2'),
	('cookie3','secret-key3'),
	('cookie4','secret-key4');

INSERT INTO public.urls (short,long,cookie,deleted) VALUES
	('abcdABCD','http://example.org','cookie4',false),
	('ABCDabcd','http://example2.org','cookie2',false),
	('abcdabcd','http://example3.org','cookie2',false),
	('12345678','http://test1.org','cookie1',false),
	('87654321','http://test2.org','cookie1',false),
	('AAAAAAAA','http://sample1.org','cookie3',false);
`)
	require.NoError(t, err)
	err = s.Close()
	require.NoError(t, err)
}

func TestIntegrationOpen(t *testing.T) {
	db := postgres{conn: testDSN}
	err := db.open()
	require.NoError(t, err)
}

func TestIntegrationCreate(t *testing.T) {
	db := postgres{conn: testDSN}
	var err error
	db.db, err = sql.Open("postgres", db.conn)
	require.NoError(t, err)
	err = db.create()
	require.NoError(t, err)
}

// Close() error
func TestIntegrationClose(t *testing.T) {
	db := postgres{conn: testDSN}
	err := db.open()
	require.NoError(t, err)
	err = db.Close()
	require.NoError(t, err)
	err = db.Ping()
	require.Error(t, err)
}

// Ping() error
func TestDBIntegrationPing(t *testing.T) {
	db, err := NewPostgreSQL(testDSN)
	require.NoError(t, err)
	err = db.Ping()
	require.NoError(t, err)
}

// ReadByCookie(string) (models.ClientData, error)
func TestDBIntegrationReadByCookie(t *testing.T) {
	dbPreparation(t)
	db, err := NewPostgreSQL(testDSN)
	require.NoError(t, err)
	e := models.ClientData{Cookie: "cookie2", Key: "secret-key2", Short: []models.ShortData{{Short: "ABCDabcd", Long: "http://example2.org"}, {Short: "abcdabcd", Long: "http://example3.org"}}}
	data, err := db.ReadByCookie("cookie2")
	require.NoError(t, err)
	require.Equal(t, e, data)
}

// ReadByTag(string) (models.ShortData, error)
func TestDBIntegrationReadByTag(t *testing.T) {
	dbPreparation(t)
	db, err := NewPostgreSQL(testDSN)
	require.NoError(t, err)
	expected := models.ShortData{Short: "ABCDabcd", Long: "http://example2.org"}
	data, err := db.ReadByTag("ABCDabcd")
	require.NoError(t, err)
	require.Equal(t, expected, data)
}

// TagByURL(string, string) (string, error)
func TestDBIntegrationTagByURL(t *testing.T) {
	dbPreparation(t)
	db, err := NewPostgreSQL(testDSN)
	require.NoError(t, err)
	expected := "ABCDabcd"
	data, err := db.TagByURL("http://example2.org", "cookie2")
	require.NoError(t, err)
	require.Equal(t, expected, data)
}

// Write(models.ClientData) error
func TestDBIntegrationWrite(t *testing.T) {
	dbPreparation(t)
	db, err := NewPostgreSQL(testDSN)
	require.NoError(t, err)
	value := models.ClientData{Cookie: "cookie2", Key: "secret-key2", Short: []models.ShortData{{Short: "AbCdAbCd", Long: "http://example4.org"}}}
	expected := models.ClientData{Cookie: "cookie2", Key: "secret-key2", Short: []models.ShortData{{Short: "ABCDabcd", Long: "http://example2.org"}, {Short: "abcdabcd", Long: "http://example3.org"}, {Short: "AbCdAbCd", Long: "http://example4.org"}}}
	err = db.Write(value)
	require.NoError(t, err)
	val, err := db.ReadByCookie("cookie2")
	require.NoError(t, err)
	require.Equal(t, expected, val)
	e1 := models.ClientData{Cookie: "cookie20", Key: "secret-key20", Short: []models.ShortData{{Short: "12345679", Long: "http://example5.org"}}}
	err = db.Write(e1)
	require.NoError(t, err)
	val, err = db.ReadByCookie("cookie20")
	require.NoError(t, err)
	require.Equal(t, e1, val)
}

// Cleaner(<-chan models.DelWorker, int)
func TestDBIntegrationCleaner(t *testing.T) {
	dbPreparation(t)
	sig := make(chan os.Signal)
	wg := sync.WaitGroup{}
	inputCh := make(chan models.DelWorker)
	db, err := NewPostgreSQL(testDSN)
	require.NoError(t, err)
	db.Cleaner(sig, &wg, inputCh, 2)
	value := models.DelWorker{Cookie: "cookie3", Tags: []string{"AAAAAAAA"}}
	inputCh <- value
	time.Sleep(15 * time.Second)
	expected := models.ClientData{Cookie: "cookie3", Key: "secret-key3", Short: []models.ShortData{{Short: "AAAAAAAA", Long: "http://sample1.org", Deleted: true}}}
	val, err := db.ReadByCookie("cookie3")
	require.NoError(t, err)
	require.Equal(t, expected, val)

}

/*
INSERT INTO public.ids (cookie,"key") VALUES
	 ('cookie1','secret-key1'),
	 ('cookie2','secret-key2'),
	 ('cookie3','secret-key3'),
	 ('cookie4','secret-key4');

INSERT INTO public.urls (short,long,cookie,deleted) VALUES
	 ('abcdABCD','http://example.org','cookie4',false),
	 ('ABCDabcd','http://example2.org','cookie2',false),
	 ('abcdabcd','http://example3.org','cookie2',false),
	 ('12345678','http://test1.org','cookie1',false),
	 ('87654321','http://test2.org','cookie1',false),
	 ('AAAAAAAA','http://sample1.org','cookie3',false);
*/

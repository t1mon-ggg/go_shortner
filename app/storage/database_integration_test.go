package storage

import (
	"context"
	"database/sql"
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
	db, err := NewPostgreSQL(testDSN)
	require.NoError(t, err)
	type args struct {
		done    <-chan struct{}
		wg      *sync.WaitGroup
		inputCh chan models.DelWorker
		workers int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "start cleaner",
			args: args{
				done:    make(<-chan struct{}),
				wg:      &sync.WaitGroup{},
				inputCh: make(chan models.DelWorker),
				workers: 10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.Cleaner(tt.args.done, tt.args.wg, tt.args.inputCh, tt.args.workers)
			tt.args.inputCh <- models.DelWorker{Cookie: "cookie3", Tags: []string{"AAAAAAAA"}}
			time.Sleep(5 * time.Second)
			data, _ := db.ReadByTag("AAAAAAAA")
			require.True(t, data.Deleted)
		})
	}
}
func TestDBIntegrationGetStats(t *testing.T) {
	dbPreparation(t)
	db, err := NewPostgreSQL(testDSN)
	require.NoError(t, err)
	expected := models.Stats{URLs: 6, Users: 4}
	data, err := db.GetStats()
	require.NoError(t, err)
	require.Equal(t, expected, data)
	t.Log(data)
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

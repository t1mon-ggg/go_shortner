package storage

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/t1mon-ggg/go_shortner/app/helpers"
	"github.com/t1mon-ggg/go_shortner/app/models"
	"github.com/t1mon-ggg/go_shortner/app/storage/internal/mock_storage"
)

var (
	ctrl    *gomock.Controller
	s       *mock_storage.MockStorage
	t       *testing.T
	inputCh chan models.DelWorker
)

func newWorker(t *testing.T, input <-chan models.DelWorker) {
	for task := range input {
		t.Log("recivied task:", task)
	}
}

func init() {
	t = &testing.T{}

	inputCh = make(chan models.DelWorker)
	fanOutChs := helpers.FanOut(inputCh, 2)
	for _, fanOutCh := range fanOutChs {
		go newWorker(t, fanOutCh)
	}

	ctrl = gomock.NewController(t)
	s = mock_storage.NewMockStorage(ctrl)
	go func() {
		defer ctrl.Finish()
	}()

	s.EXPECT().Ping().Return(nil)
	s.EXPECT().ReadByCookie("cookie2").Return(
		models.ClientData{
			Cookie: "cookie2",
			Key:    "secret-key2",
			Short: []models.ShortData{
				{
					Short: "ABCDabcd",
					Long:  "http://example2.org",
				},
				{
					Short: "abcdabcd",
					Long:  "http://example3.org",
				},
			},
		},
		nil,
	)
	s.EXPECT().ReadByCookie("cookie22").Return(
		models.ClientData{
			Cookie: "cookie22",
			Key:    "secret-key22",
			Short: []models.ShortData{
				{
					Short: "ABCDabcd",
					Long:  "http://example2.org",
				},
				{
					Short: "abcdabcd",
					Long:  "http://example3.org",
				},
				{
					Short: "AbCdAbCd",
					Long:  "http://example4.org",
				},
			},
		},
		nil,
	)
	s.EXPECT().ReadByCookie("cookie3").Return(
		models.ClientData{
			Cookie: "cookie3",
			Key:    "secret-key3",
			Short: []models.ShortData{
				{Short: "AAAAAAAA",
					Long:    "http://sample1.org",
					Deleted: true,
				},
			},
		},
		nil,
	)
	s.EXPECT().ReadByTag("ABCDabcd").Return(
		models.ShortData{
			Short: "ABCDabcd",
			Long:  "http://example2.org",
		},
		nil,
	)
	s.EXPECT().TagByURL("http://example2.org", "cookie2").Return("ABCDabcd", nil)
	s.EXPECT().Write(models.ClientData{
		Cookie: "cookie22",
		Key:    "secret-key22",
		Short: []models.ShortData{
			{
				Short: "AbCdAbCd",
				Long:  "http://example4.org",
			},
		},
	}).
		Return(nil)
	s.EXPECT().Close().Return(nil)
	s.EXPECT().Cleaner(inputCh, 2)

}

// Ping() error
func Test_Ping(t *testing.T) {
	err := s.Ping()
	require.NoError(t, err)
}

// ReadByCookie(string) (models.ClientData, error)
func Test_ReadByCookie(t *testing.T) {
	expected := models.ClientData{
		Cookie: "cookie2",
		Key:    "secret-key2",
		Short: []models.ShortData{
			{
				Short: "ABCDabcd",
				Long:  "http://example2.org",
			},
			{
				Short: "abcdabcd",
				Long:  "http://example3.org",
			},
		},
	}
	data, err := s.ReadByCookie("cookie2")
	require.NoError(t, err)
	require.Equal(t, expected, data)
}

// ReadByTag(string) (models.ShortData, error)
func Test_ReadByTag(t *testing.T) {
	expected := models.ShortData{
		Short: "ABCDabcd",
		Long:  "http://example2.org",
	}
	data, err := s.ReadByTag("ABCDabcd")
	require.NoError(t, err)
	require.Equal(t, expected, data)
}

// TagByURL(string, string) (string, error)
func Test_TagByURL(t *testing.T) {
	expected := "ABCDabcd"
	data, err := s.TagByURL("http://example2.org", "cookie2")
	require.NoError(t, err)
	require.Equal(t, expected, data)
}

// Write(models.ClientData) error
func Test_Write(t *testing.T) {
	w1 := models.ClientData{
		Cookie: "cookie22",
		Key:    "secret-key22",
		Short: []models.ShortData{
			{Short: "AbCdAbCd",
				Long: "http://example4.org",
			},
		},
	}
	e1 := models.ClientData{
		Cookie: "cookie22",
		Key:    "secret-key22",
		Short: []models.ShortData{
			{Short: "ABCDabcd",
				Long: "http://example2.org",
			},
			{Short: "abcdabcd",
				Long: "http://example3.org",
			},
			{Short: "AbCdAbCd",
				Long: "http://example4.org",
			},
		},
	}
	err := s.Write(w1)
	require.NoError(t, err)
	val, err := s.ReadByCookie("cookie22")
	require.NoError(t, err)
	require.Equal(t, e1, val)
}

// Cleaner(<-chan models.DelWorker, int)
func Test_Cleaner(t *testing.T) {
	s.Cleaner(inputCh, 2)
	value := models.DelWorker{Cookie: "cookie3", Tags: []string{"AAAAAAAA"}}
	inputCh <- value
	time.Sleep(15 * time.Second)
	expected := models.ClientData{Cookie: "cookie3", Key: "secret-key3", Short: []models.ShortData{{Short: "AAAAAAAA", Long: "http://sample1.org", Deleted: true}}}
	val, err := s.ReadByCookie("cookie3")
	require.NoError(t, err)
	require.Equal(t, expected, val)
}

func Test_Close(t *testing.T) {
	err := s.Close()
	require.NoError(t, err)
}
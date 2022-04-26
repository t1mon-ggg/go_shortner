package storage

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/t1mon-ggg/go_shortner/internal/app/mock"
	. "github.com/t1mon-ggg/go_shortner/internal/app/models"
)

func Test_DB_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockDatabase(ctrl)

	db.EXPECT().Ping().Return(nil)

	err := db.Ping()
	require.NoError(t, err)
}

func Test_DB_Write(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockDatabase(ctrl)
	s1 := make(map[string]WebData)
	s2 := make(map[string]WebData)
	s1["cookie2"] = WebData{
		Key: "secret-key2",
		Short: map[string]string{
			"ABCDabcd": "http://example2.org",
			"abcdabcd": "http://example3.org",
			"AbCdAbCD": "http://example4.org",
		},
	}
	s2["cookie3"] = WebData{
		Key: "secret-key3",
		Short: map[string]string{
			"AbCdAbCd": "http://example4.org",
		},
	}
	db.EXPECT().Write(s1).Return(nil)
	db.EXPECT().Write(s2).Return(fmt.Errorf("not unique url"))

	err := db.Write(s1)
	require.NoError(t, err)
	err = db.Write(s2)
	require.Error(t, err)

}

func Test_DB_ReadByCookie(t *testing.T) {}

func Test_DB_ReadByTag(t *testing.T) {}

func Test_DB_TagByURL(t *testing.T) {}

func Test_DB_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockDatabase(ctrl)
	db.EXPECT().Close().Return(nil)

	err := db.Close()
	require.NoError(t, err)

}

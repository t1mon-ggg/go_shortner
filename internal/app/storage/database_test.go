package storage

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/t1mon-ggg/go_shortner/internal/app/mock"
)

func Test_DB_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockDatabase(ctrl)

	db.EXPECT().Ping().Return(nil)

	err := db.Ping()
	require.NoError(t, err)
}

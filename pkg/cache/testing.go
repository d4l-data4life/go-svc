package cache

import (
	"github.com/go-redis/redismock/v8"
)

func InitializeRedisMock() redismock.ClientMock {
	db, mock := redismock.NewClientMock()
	rdb = db
	return mock
}

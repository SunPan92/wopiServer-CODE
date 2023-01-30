package cache

import (
	"github.com/patrickmn/go-cache"
	"sync"
	"time"
	"wopi-server/g"
)

var (
	once sync.Once
)

//InitCache 初始化本地缓存
func InitCache() {
	once.Do(func() {
		g.Cache = cache.New(60*time.Minute, 12*time.Hour)
	})
}

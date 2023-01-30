package g

import (
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

//Log 全局日志
var Log *zap.Logger

//Cache 全局缓存
var Cache *cache.Cache

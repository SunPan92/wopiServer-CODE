package middleware

import (
	"wopi-server/common"

	"github.com/kataras/iris/v12"

	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

// IrisRecovery recover 项目可能出现的panic
func IrisRecovery(stack bool) iris.Handler {
	return func(ctx iris.Context) {
		defer func() {
			if err := recover(); err != nil {
				stacktrace := getStackErrorMsg()
				// when stack finishes
				stackMsg := fmt.Sprintf("从错误中回复：('%s')\n", ctx.HandlerName())
				stackMsg += fmt.Sprintf("\n%s", stacktrace)

				//check for a broken connection, as it is not really a condition that warrants a panic stack here
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}
				request, _ := httputil.DumpRequest(ctx.Request(), false)
				logError(stack, ctx.Request().URL.Path, string(request), err, stackMsg)
				if brokenPipe {
					return
				} else {
					ctx.StatusCode(http.StatusOK)
					_, _ = ctx.JSON(common.GenFailedMsg(ErrorString(err)))
				}
				ctx.StopExecution()
			}
		}()
		ctx.Next()
	}
}

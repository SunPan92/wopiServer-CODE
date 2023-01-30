package main

import (
	"go.uber.org/zap"
	"wopi-server/config"
	"wopi-server/g"
	"wopi-server/handler"
	"wopi-server/middleware"

	"github.com/kataras/iris/v12"

	"sync"
)

var once sync.Once

// InitRouter init iris router engine
func InitRouter() *iris.Application {
	var app *iris.Application
	once.Do(func() {
		app = iris.New()
		app.Use(middleware.IrisRecovery(false))
		app.Use(middleware.IrisLogger())
		app.HandleDir(config.Bean.Server.Context, "./static")
		// web的上下文

		context := app.Party(config.Bean.Server.Context)
		wopi := context.Party("/wopi")
		{
			wopi.Get("/files/{id}", handler.LocalFileHandler.GetFileInfo)
			wopi.Get("/files/{id}/contents", handler.LocalFileHandler.GetFileContent)
			wopi.Post("/files/{id}/contents", handler.LocalFileHandler.PutFile)
			wopi.Post("/files/{id}", func(ctx iris.Context) { ctx.StatusCode(200) })
			wopi.Get("/collaboraUrl", handler.GetCollaboraUrl)
		}
		context.Put("/removeExpiredFile", handler.LocalFileHandler.RemoveFile)
		context.Get("/files/history", handler.LocalFileHandler.GetHistoryFile)
		context.Get("/files/rollback", handler.LocalFileHandler.Rollback)
		context.Get("/ping", func(ctx iris.Context) {
			_, err := ctx.WriteString("pong")
			if err != nil {
				g.Log.Error("", zap.Any("", err))
				return
			}
		})
	})
	return app
}

package handler

import (
	"fmt"
	"github.com/duke-git/lancet/datetime"
	"github.com/duke-git/lancet/fileutil"
	"github.com/kataras/iris/v12"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"strconv"
	"time"
	"wopi-server/common"
	"wopi-server/common/util"
	"wopi-server/g"

	"strings"
	"wopi-server/config"
	"wopi-server/table"
	"wopi-server/third/db"
)

// LocalFileRootdir 本地文件的根目录
var LocalFileRootdir string

func Initial() {
	LocalFileRootdir = config.Bean.Server.LocalFileRootdir
}

type remoteFile struct {
}

var RemoteFile = remoteFile{}

//	GetFileInfo 获取文件信息
func (remoteFile) GetFileInfo(ctx iris.Context) {
	id, err := ctx.Params().GetInt64("id")
	if err != nil {
		g.Log.Error("读取id错误，", zap.Any("", err))
		errMsg, _ := fmt.Printf("错误的请求参数，id = %v", id)
		_, _ = ctx.JSON(common.H{`errMsg`: errMsg})
		return
	}
	token := ctx.Params().Get("access_token")
	resp := common.H{
		`UserCanWrite`:     len(token) > 0,
		`UserFriendlyName`: "user",
	}
	var f table.FileInfo
	pgDb := db.PgDb
	pgDb.Model(&table.FileInfo{}).First(&f, id)
	if len(f.FileName) == 0 {
		f.FileName = strconv.FormatInt(id, 10) + ".docx"
		f.FileSize = 0
	}
	resp[`BaseFileName`] = f.FileName
	resp[`Size`] = f.FileSize
	resp[`UserId`] = f.CreateUser
	_, _ = ctx.JSON(resp)
}

// GetFileContent 获取文件内容
func (remoteFile) GetFileContent(ctx iris.Context) {
	id, err := ctx.Params().GetInt64("id")
	if err != nil {
		errMsg, _ := fmt.Printf("错误的请求参数，id = %v", id)
		ctx.JSON(common.H{"errMsg": errMsg})
		return
	}
	var f table.FileInfo
	db.PgDb.Model(&table.FileInfo{}).First(&f, id)
	if len(f.FileName) == 0 {
		ctx.Write(make([]byte, 0))
		return
	}
	fapStr := LocalFileRootdir + f.FilePath
	if fileutil.IsExist(fapStr) {
		data, err := ioutil.ReadFile(fapStr)
		if err != nil {
			g.Log.Error("read file error ", zap.Any("", err))
			ctx.JSON(common.H{"errMsg": "read file error"})
		} else {
			_, _ = ctx.Write(data)
		}
	} else {
		db.PgDb.Model(&table.FileInfo{}).Where("id = ?", id).Update("del_flag", 1)
		if _, err := ctx.JSON(common.H{"errMsg": "文件不存在，path =" + fapStr}); err != nil {
			g.Log.Error("return response error", zap.Any("", err))
		}

	}
}

// PutFile 保存文件（覆盖）
func (remoteFile) PutFile(ctx iris.Context) {
	id, err := ctx.Params().GetInt64("id")
	if err != nil {
		errMsg, _ := fmt.Printf("错误的请求参数，id = %v", id)
		_, _ = ctx.JSON(common.H{"errMsg": errMsg})
		return
	}
	var f table.FileInfo
	db.PgDb.Model(&table.FileInfo{}).First(&f, id)
	if len(f.FileName) == 0 {
		f.FilePath = "/" + strconv.FormatInt(id, 10) + ".docx"
		_, err := os.Create(LocalFileRootdir + f.FilePath)
		if err != nil {
			g.Log.Error("", zap.Any("", err))
			return
		}
	}
	if data, err := ioutil.ReadAll(ctx.Request().Body); err != nil {
		g.Log.Error("read request body error", zap.Any("", err))
		_, _ = ctx.JSON(common.H{"errMsg": "read request body error"})
	} else {
		fapStr := LocalFileRootdir + f.FilePath
		if writeLocalFile(fapStr, data) {
			ctx.JSON(common.H{"errMsg": "write file error"})
		} else {
			ctx.StatusCode(200)
		}

	}
}

// writeLocalFile 将数据写入本地并同时备份
func writeLocalFile(filepath string, data []byte) bool {
	filename := util.File.GetFilename(filepath)
	parentDir := util.File.GetParentDir(filepath)
	sourceFile := parentDir + "source_" + filename
	// 备份文件
	if !fileutil.IsExist(sourceFile) {
		err := fileutil.CopyFile(filepath, sourceFile)
		if err != nil {
			g.Log.Error("copy file error ", zap.Any("", err))
			return false
		}
	} else {
		timeToStr := datetime.FormatTimeToStr(time.Now(), "yyyy-mm-dd hh:mm:ss")
		tStr := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(timeToStr, "-", ""), ":", ""), " ", "")
		bakFile := parentDir + tStr + "_" + filename
		bak := util.File.Write(bakFile, data)
		if !bak {
			g.Log.Error("备份文件失败", zap.Any("back file = ", bakFile))
			return false
		}
	}
	return util.File.Write(filepath, data)
}

package handler

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"wopi-server/common"
	"wopi-server/common/util"
	"wopi-server/config"
	"wopi-server/g"

	"github.com/duke-git/lancet/datetime"
	"github.com/duke-git/lancet/fileutil"
	"github.com/kataras/iris/v12"
	"go.uber.org/zap"
)

type localFile struct {
}

var LocalFileHandler = localFile{}

//	GetFileInfo 获取文件信息
func (localFile) GetFileInfo(ctx iris.Context) {
	info := getFileInfo(ctx)
	if info == nil {
		ctx.StatusCode(405)
		if _, err := ctx.WriteString("获取不到文件信息"); err != nil {
			g.Log.Error("return response error", zap.Any("", err))
		}
		return
	}
	filepath := info.Filepath
	fileAbsolutePath := LocalFileRootdir + filepath
	var LastModifiedTime time.Time
	if util.File.IsExist(fileAbsolutePath) {
		f, err := os.Open(fileAbsolutePath)
		if err == nil {
			defer f.Close()
			if fi, err := f.Stat(); err != nil {
				g.Log.Warn("stat fileinfo error")
				LastModifiedTime = time.Now()
			} else {
				LastModifiedTime = fi.ModTime()
			}
		} else {
			g.Log.Warn("打开文件错误，path = "+fileAbsolutePath, zap.Any("", err))
		}
	} else {
		LastModifiedTime = time.Now()
	}
	timeStr := LastModifiedTime.Format(util.ISO_8601)
	token := ctx.URLParam("access_token")
	resp := common.H{
		`UserCanWrite`:                       len(token) > 0 && config.Bean.Server.Token == token,
		`UserFriendlyName`:                   "user",
		`DisableBrowserCachingOfUserContent`: true,
		`BaseFileName`:                       info.Filename,
		`LastModifiedTime`:                   timeStr,
	}
	if _, err := ctx.JSON(resp); err != nil {
		g.Log.Error("return response error", zap.Any("", err))
	}
}

// GetFileContent 获取文件内容
func (localFile) GetFileContent(ctx iris.Context) {
	info := getFileInfo(ctx)
	if info == nil {
		ctx.StatusCode(405)
		if _, err := ctx.WriteString("获取不到文件信息"); err != nil {
			g.Log.Error("return response error", zap.Any("", err))
		}
		return
	}
	fapStr := LocalFileRootdir + info.Filepath
	if util.File.IsExist(fapStr) {
		data, err := ioutil.ReadFile(fapStr)
		if err != nil {
			g.Log.Error("read file error ", zap.Any("", err))
			ctx.StatusCode(405)
			if _, err := ctx.WriteString("read file error"); err != nil {
				g.Log.Error("return response error", zap.Any("", err))
			}
		} else {
			if _, err := ctx.Write(data); err != nil {
				g.Log.Error("return response error", zap.Any("", err))
			}
		}
	} else {
		if b := util.File.MkAll(fapStr); !b {
			g.Log.Error("创建本地文件失败")
		}
		var data []byte
		if _, err := ctx.Write(data); err != nil {
			g.Log.Error("return response error", zap.Any("", err))
		}
	}
}

// PutFile 保存文件（覆盖）
func (localFile) PutFile(ctx iris.Context) {
	info := getFileInfo(ctx)
	if info == nil {
		ctx.StatusCode(405)
		if _, err1 := ctx.WriteString("获取不到文件信息"); err1 != nil {
			g.Log.Error("return response error", zap.Any("", err1))
		}
		return
	}
	fapStr := LocalFileRootdir + info.Filepath
	if util.File.IsNotExist(fapStr) {
		if b := util.File.MkAll(fapStr); !b {
			g.Log.Error("创建本地文件失败")
			ctx.StatusCode(500)
			if _, err2 := ctx.WriteString("创建本地文件失败"); err2 != nil {
				g.Log.Error("return response error", zap.Any("", err2))
			}
			return
		}
	}
	if data, err3 := ioutil.ReadAll(ctx.Request().Body); err3 != nil {
		g.Log.Error("read request body error", zap.Any("", err3))
		if _, err4 := ctx.WriteString("read request body error"); err4 != nil {
			g.Log.Error("return response error", zap.Any("", err4))
		}
	} else {
		if writeLocalFile(fapStr, data) {
			ctx.StatusCode(200)
		} else {
			ctx.StatusCode(500)
			if _, err5 := ctx.WriteString("写本地文件失败"); err5 != nil {
				g.Log.Error("return response error", zap.Any("", err5))
			}
		}
	}
}

//RemoveFile 清理本地过期的备份文件
func (f localFile) RemoveFile(ctx iris.Context) {
	files, err := util.File.GetAllFile(LocalFileRootdir)
	if err != nil {
		g.Log.Error("获取本地根目录的所有文件失败", zap.Any("", err))
		ctx.StatusCode(500)
		return
	}
	date := ctx.URLParamTrim("date")
	if len(date) == 0 {
		defaultDate := time.Now().AddDate(0, 0, -30)
		date = datetime.FormatTimeToStr(defaultDate, "yyyy-mm-dd")
	}
	date = strings.ReplaceAll(date, "-", "")
	t, err := strconv.Atoi(date)
	if err != nil {
		g.Log.Error("错误的日期参数", zap.Any("", err))
		if _, err := ctx.WriteString("错误的日期参数：" + date); err != nil {
			g.Log.Error("return response error", zap.Any("", err))
		}
		return
	}
	var success = true
	timeRex := regexp.MustCompile(`^\d{4}[0,1]\d[0-3]\d[0-2]\d[0-5]\d[0-5]\d_`)
	for _, f := range files {
		filename := util.File.GetFilename(f)
		//根据规则提取关键信息
		timeStr := timeRex.FindString(filename)
		if len(timeStr) == 0 {
			continue
		}
		tn, err := strconv.Atoi(timeStr[:8])
		if err != nil {
			g.Log.Error("", zap.Any("", err))
			success = false
			continue
		}
		if t > tn {
			if err := fileutil.RemoveFile(f); err != nil {
				g.Log.Warn("删除文件失败", zap.Any("", err))
				success = false
			} else {
				g.Log.Info("删除文件成功，文件 = " + f)
			}
		}
	}
	if _, err := ctx.WriteString("清除历史文件：" + fmt.Sprintf("%v", success)); err != nil {
		g.Log.Error("return response error", zap.Any("", err))
	}
}

//GetHistoryFile 查询文件的历史保存记录
func (f localFile) GetHistoryFile(ctx iris.Context) {
	filepath := ctx.URLParamTrim("filepath")
	if len(filepath) == 0 {
		ctx.JSON(common.H{"errMsg": "缺少filepath参数"})
		return
	}
	fullpath := LocalFileRootdir + filepath
	parentDir := util.File.GetParentDir(fullpath)
	filename := util.File.GetFilename(fullpath)
	i := strings.LastIndex(filepath, "/")
	fs, err := fileutil.ListFileNames(parentDir)
	if err != nil {
		g.Log.Error("读取目录下的文件失败", zap.Any("", err))
		ctx.JSON(common.H{"errMsg": "读取目录下的文件失败"})
		return
	}
	fileReg := regexp.MustCompile(`^\d{4}[0,1]\d[0-3]\d[0-2]\d[0-5]\d[0-5]\d_` + filename)
	var res fileList
	for _, f := range fs {
		historyFile := fileReg.FindString(f)
		if len(historyFile) == 0 {
			continue
		}
		modifyTime, err := strconv.ParseInt(f[0:14], 10, 64)
		if err != nil {
			g.Log.Error("获取修改时间错误", zap.Any("", err))
			ctx.StatusCode(500)
			return
		}
		res = append(res, FileInfo{
			Filename:      filename,
			Filepath:      filepath[:i] + "/" + f,
			ModifyTime:    f[0:4] + "年" + f[4:6] + "月" + f[6:8] + "日 " + f[8:10] + ":" + f[10:12] + ":" + f[12:14],
			ModifyTimeNum: modifyTime,
		})
	}
	sort.Sort(res)
	_, _ = ctx.JSON(res)
}

// Rollback 回滚文件到历史保存的记录
func (f localFile) Rollback(ctx iris.Context) {
	filepath := ctx.URLParamTrim("filepath")
	if len(filepath) == 0 {
		ctx.JSON(common.H{"errMsg": "缺少filepath参数"})
		return
	}
	historyFilepath := LocalFileRootdir + filepath
	parentDir := util.File.GetParentDir(historyFilepath)
	filename := util.File.GetFilename(historyFilepath)
	dstFile := parentDir + "/" + filename[15:]
	err := fileutil.CopyFile(historyFilepath, dstFile)
	if err != nil {
		g.Log.Error("复制原始文件错误", zap.Any("", err))
		ctx.StatusCode(500)
	} else {
		ctx.WriteString("ok")
	}
}

type fileList []FileInfo

// 实现sort SDK 中的Interface接口

func (fs fileList) Len() int {
	//返回传入数据的总数
	return len(fs)
}
func (fs fileList) Swap(i, j int) {
	//两个对象满足Less()则位置对换
	//表示执行交换数组中下标为i的数据和下标为j的数据
	fs[i], fs[j] = fs[j], fs[i]
}

func (fs fileList) Less(i, j int) bool {
	//按字段比较大小,此处是降序排序
	//返回数组中下标为i的数据是否小于下标为j的数据
	return fs[i].ModifyTimeNum > fs[j].ModifyTimeNum
}

type FileInfo struct {
	Id            string
	Filename      string
	Filepath      string
	ModifyTime    string
	ModifyTimeNum int64
}

func getFileInfo(ctx iris.Context) *FileInfo {
	var res FileInfo
	id := ctx.Params().Get("id")
	if len(id) == 0 {
		errMsg := "缺少请求参数[id]"
		_, _ = ctx.JSON(common.H{`errMsg`: errMsg})
		return nil
	}
	v, b := g.Cache.Get(id)
	if b {
		m, err := util.ObjUtil.Obj2map(v)
		if err != nil {
			g.Log.Error("转换缓存对象错误", zap.Any("", err))
			ctx.StatusCode(500)
		} else {
			res.Id = id
			res.Filename = fmt.Sprintf("%v", m["filename"])
			res.Filepath = fmt.Sprintf("%v", m["filepath"])
			return &res
		}
	} else {
		ctx.WriteString(fmt.Sprintf("id = %v 的文件信息不存在", id))
	}
	return nil
}

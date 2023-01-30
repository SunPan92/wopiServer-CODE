package handler

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"github.com/kataras/iris/v12"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"wopi-server/common"
	"wopi-server/common/util"
	"wopi-server/config"
	"wopi-server/g"
)

const key = "wopiClientUrl"

type ProofKey struct {
	Exponent    string `xml:"exponent,attr"`
	OldExponent string `xml:"oldexponent,attr"`
	Value       string `xml:"value,attr"`
	OldValue    string `xml:"oldvalue,attr"`
	Modulus     string `xml:"modulus,attr"`
	OldModulus  string `xml:"oldmodulus,attr"`
}

type Action struct {
	Default bool   `xml:"default,attr"`
	Ext     string `xml:"ext,attr"`
	Name    string `xml:"name,attr"`
	Urlsrc  string `xml:"urlsrc,attr"`
}

type App struct {
	Name       string   `xml:"name,attr"`
	FavIconUrl string   `xml:"favIconUrl,attr"`
	Actions    []Action `xml:"action"`
}

type NetZone struct {
	Name string `xml:"name,attr"`
	Apps []App  `xml:"app"`
}

type WopiDiscovery struct {
	XMLName  xml.Name `xml:"wopi-discovery"` // 指定最外层的标签
	NetZone  NetZone  `xml:"net-zone"`
	ProofKey ProofKey `xml:"proof-key"`
}

func GetCollaboraUrl(ctx iris.Context) {
	filepath := ctx.URLParamTrim("filepath")
	if len(filepath) == 0 {
		ctx.JSON(common.H{"errMsg": "缺少filepath参数"})
		return
	}
	wopiClientUrl := getWopiEditUrl()
	if len(wopiClientUrl) == 0 {
		ctx.JSON(common.H{"errMsg": "获取wopi hosting discovery错误"})
		return
	}
	fileId := StringMd5(filepath)
	fileInfo := common.H{
		"id":       fileId,
		"filepath": filepath,
		"filename": util.File.GetFilename(filepath),
	}
	g.Cache.SetDefault(fileId, fileInfo)
	host := ctx.Request().Host
	ctx.JSON(common.H{
		"url":      wopiClientUrl,
		"wopiHost": "http://" + host + config.Bean.Server.Context + "/wopi/files",
		"fileId":   fileId,
	})
}

// StringMd5 计算字符串的md5值
func StringMd5(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// StringBase64 计算字符串的Base64值
func StringBase64(text string) string {
	txt := []byte(text)
	return base64.StdEncoding.EncodeToString(txt)
}

func getWopiEditUrl() string {
	value, found := g.Cache.Get(key)
	if found {
		return value.(string)
	}
	var wopiClientUrl string
	wopiHosting := config.Bean.Server.WopiServer + "/hosting/discovery"
	resp, err1 := http.Get(wopiHosting)
	if err1 != nil {
		g.Log.Error("获取wopi hosting 失败", zap.Any("", err1))
		return wopiClientUrl
	}
	data, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		g.Log.Error("获取响应数据失败", zap.Any("", err2))
		return wopiClientUrl
	}
	wopiDis := WopiDiscovery{}
	if err3 := xml.Unmarshal(data, &wopiDis); err3 != nil {
		g.Log.Error("解析wopi xml 失败", zap.Any("", err3))
		return wopiClientUrl
	}
	wopiClientUrl = wopiDis.NetZone.Apps[0].Actions[0].Urlsrc
	if len(wopiClientUrl) > 0 {
		g.Cache.SetDefault(key, wopiClientUrl)
	}
	return wopiClientUrl
}

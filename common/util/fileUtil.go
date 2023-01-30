package util

import (
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"strings"
	"wopi-server/g"
)

type fileUtil struct {
}

var File = fileUtil{}

//IsExist 判断文件或文件夹是否存在
func (fileUtil) IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		fmt.Println(err)
		return false
	}
	return true
}

//IsNotExist 判断文件或文件夹是否存在
func (fileUtil) IsNotExist(path string) bool {
	return !File.IsExist(path)
}

func (fileUtil) MkAll(path string) bool {
	dir := File.GetParentDir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Printf(err.Error())
		return false
	}
	file, err2 := os.Create(path)
	if err2 != nil {
		fmt.Printf(err2.Error())
		return false
	}
	defer file.Close()
	return true
}

//GetFilename 根据文件路径获取文件名
func (fileUtil) GetFilename(filepath string) string {
	i := strings.LastIndex(filepath, "/")
	return filepath[i+1:]
}

//GetParentDir 根据文件路径获取文件的父文件夹
func (fileUtil) GetParentDir(filepath string) string {
	i := strings.LastIndex(filepath, "/")
	return filepath[0 : i+1]
}

//Write 向目标路径写文件
func (fileUtil) Write(dst string, data []byte) bool {
	if err := ioutil.WriteFile(dst, data, os.ModePerm); err != nil {
		g.Log.Error("write file error ", zap.Any("", err))
		return false
	}
	return true
}

//GetAllFile 递归获取指定目录下的所有文件名
func (fileUtil) GetAllFile(pathname string) ([]string, error) {
	var result []string
	fis, err := ioutil.ReadDir(pathname)
	if err != nil {
		fmt.Printf("读取文件目录失败，pathname=%v, err=%v \n", pathname, err)
		return result, err
	}
	// 所有文件/文件夹
	for _, fi := range fis {
		fullname := pathname + "/" + fi.Name()
		// 是文件夹则递归进入获取;是文件，则压入数组
		if fi.IsDir() {
			temp, err := File.GetAllFile(fullname)
			if err != nil {
				fmt.Printf("读取文件目录失败,fullname=%v, err=%v", fullname, err)
				return result, err
			}
			result = append(result, temp...)
		} else {
			result = append(result, fullname)
		}
	}
	return result, nil
}

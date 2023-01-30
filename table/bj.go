package table

import (
	"wopi-server/common"
)

// FileInfo mapped from table <file_info>
type FileInfo struct {
	ID             int64       `gorm:"column:id;type:int8;primaryKey"                json:"id"`             // 主键
	FileName       string      `gorm:"column:file_name;type:text;not null"           json:"fileName"`       // 文件名
	FilePath       string      `gorm:"column:file_path;type:text"                    json:"filePath"`       // 文件路径（真实路径）
	FileType       string      `gorm:"column:file_type;type:text"                    json:"fileType"`       // 文件类型：jpg、png、txt、doc
	PathTree       string      `gorm:"column:path_tree;type:ltree"                   json:"pathTree"`       // 文件的逻辑路径
	FileSize       int32       `gorm:"column:file_size;type:int4"                    json:"fileSize"`       // 文件大小，单位：byte
	FileMd5        string      `gorm:"column:file_md5;type:text"                     json:"fileMd5"`        // 文件md5
	CreateUser     string      `gorm:"column:create_user;type:text"                  json:"createUser"`     // 上传用户
	UploadTime     common.Time `gorm:"column:upload_time;type:timestamp"             json:"uploadTime"`     // 上传时间
	AccessTime     common.Time `gorm:"column:access_time;type:timestamp"             json:"accessTime"`     // 最后访问时间
	FilePlatformID int64       `gorm:"column:file_platform_id;type:int8"             json:"filePlatformId"` // 平台ID
	UpdateTime     common.Time `gorm:"column:update_time;type:timestamp"             json:"updateTime"`     // 更新时间
	DelFlag        int32       `gorm:"column:del_flag;type:int4"                     json:"delFlag"`        // 删除标记:0-未删除，1-已删除
	DelTime        common.Time `gorm:"column:del_time;type:timestamp"                json:"delTime"`        // 删除时间
}

// TableName FileInfo's table name
func (*FileInfo) TableName() string {
	return "file_info"
}

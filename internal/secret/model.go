package secret

import "time"

// Secret 全局密钥，用于工作流模板中 {{secret.xxx}} 引用
type Secret struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Key       string    `json:"key" gorm:"size:255;uniqueIndex;not null"`
	Value     string    `json:"value" gorm:"type:text;not null"`
	Desc      string    `json:"desc" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *Secret) TableName() string {
	return "secrets"
}

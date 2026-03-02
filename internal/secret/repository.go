package secret

import (
	"context"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) AutoMigrate() error {
	return r.db.AutoMigrate(&Secret{})
}

func (r *Repository) Create(ctx context.Context, s *Secret) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *Repository) List(ctx context.Context) ([]Secret, error) {
	var secrets []Secret
	err := r.db.WithContext(ctx).Order("`key` ASC").Find(&secrets).Error
	return secrets, err
}

func (r *Repository) Update(ctx context.Context, s *Secret) error {
	return r.db.WithContext(ctx).Model(s).Select("value", "desc").Updates(s).Error
}

func (r *Repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Secret{}, id).Error
}

// GetAll 获取所有密钥，返回 key->value 映射，实现 workflow.SecretStore 接口
func (r *Repository) GetAll(ctx context.Context) (map[string]any, error) {
	secrets, err := r.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string]any, len(secrets))
	for _, s := range secrets {
		result[s.Key] = s.Value
	}
	return result, nil
}

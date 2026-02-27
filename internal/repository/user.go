// Package repository 提供数据库操作相关功能
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"

	"gorm.io/gen"
)

// UserRepo 用户仓库
type UserRepo struct {
}

// FindUserInfoByUserID 根据用户ID查询用户信息
func (u *UserRepo) FindUserInfoByUserID() {

}

// FindUserInfoByPhone 根据手机号查询用户信息
func (u *UserRepo) FindUserInfoByPhone(phone string) (*model.User, error) {
	return query.User.Where(query.User.Phone.Eq(phone)).First()
}

// CreateUser 创建用户
func (u *UserRepo) CreateUser(user *model.User) (int64, error) {
	err := query.User.Create(user)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

// UpdateUser 更新用户信息
func (u *UserRepo) UpdateUser(id int64, user *model.User) (gen.ResultInfo, error) {
	info, err := query.User.Where(query.User.ID.Eq(id)).Updates(user)
	if err != nil {
		return info, err
	}
	return info, nil
}

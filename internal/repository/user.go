// Package repository 提供数据库操作相关功能
package repository

import (
	"aibuddy/internal/model"
	"aibuddy/internal/query"
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gen"
	"gorm.io/gorm/clause"
)

var tracer = func() trace.Tracer {
	tracer := otel.GetTracerProvider().Tracer(
		"aibuddy/internal/repository/user",
	)
	return tracer
}()

// UserRepo 用户仓库
type UserRepo struct {
}

// New 创建用户仓库实例
func New() *UserRepo {
	return &UserRepo{}
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

// Upsert 插入或更新用户信息
func (u *UserRepo) Upsert(ctx context.Context, user *model.User) error {
	_, span := tracer.Start(ctx, "UpsertUser")
	defer span.End()

	if err := query.User.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]any{
				query.User.OpenID.ColumnName().String():   user.OpenID,
				query.User.Nickname.ColumnName().String(): user.Nickname,
				query.User.Avatar.ColumnName().String():   user.Avatar,
			}),
		},
	).Create(user); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", err.Error()))
		span.SetAttributes(attribute.String("user", user.String()))
		return err
	}
	return nil
}

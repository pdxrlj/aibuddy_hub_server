// Package chatmilvus 提供 Milvus 连接和操作
package chatmilvus

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = func() trace.Tracer {
	tracer := otel.GetTracerProvider().Tracer(
		"aibuddy/pkg/milvus",
	)
	return tracer
}()

// Conn Milvus 连接
type Conn struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	_client  *milvusclient.Client
}

// New 创建 Milvus 连接实例
func New(host string, port int, username string, password string, database string) *Conn {
	return &Conn{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Database: database,
	}
}

// Connect 连接到 Milvus 服务器
func (c *Conn) Connect(ctx context.Context) (*Conn, error) {
	client, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address:  fmt.Sprintf("%s:%d", c.Host, c.Port),
		Username: c.Username,
		Password: c.Password,
		DBName:   c.Database,
	})
	if err != nil {
		return nil, err
	}
	c._client = client
	return c, nil
}

// Close 关闭连接
func (c *Conn) Close(ctx context.Context) error {
	return c._client.Close(ctx)
}

// AutoMigrate 创建向量集合（如果不存在）
// dimension: 向量维度，默认 768
func (c *Conn) AutoMigrate(ctx context.Context, collectionName string, dimension int) error {
	ctx, span := tracer.Start(ctx, "AutoMigrate")
	defer span.End()

	// 检查集合是否存在
	has, err := c._client.HasCollection(ctx, milvusclient.NewHasCollectionOption(collectionName))
	if err != nil {
		span.RecordError(err)
		return err
	}
	if has {
		span.SetAttributes(attribute.String("collectionName", collectionName), attribute.String("status", "already_exists"))
		return nil // 已存在则跳过
	}

	// 定义集合 Schema
	schema := entity.NewSchema().
		WithField(entity.NewField().
			WithName("id").
			WithDataType(entity.FieldTypeInt64).
			WithIsPrimaryKey(true).
			WithIsAutoID(true),
		).
		WithField(entity.NewField().
			WithName("query").
			WithDataType(entity.FieldTypeVarChar).
			WithTypeParams(entity.TypeParamMaxLength, "65535"),
		).
		WithField(entity.NewField().
			WithName("answer").
			WithDataType(entity.FieldTypeVarChar).
			WithTypeParams(entity.TypeParamMaxLength, "65535"),
		).
		WithField(entity.NewField().
			WithName("vector").
			WithDataType(entity.FieldTypeFloatVector).
			WithTypeParams(entity.TypeParamDim, fmt.Sprintf("%d", dimension)),
		).
		WithField(entity.NewField().
			WithName("device_id").
			WithDataType(entity.FieldTypeVarChar).
			WithTypeParams(entity.TypeParamMaxLength, "100"),
		).
		WithField(entity.NewField().
			WithName("metadata").
			WithDataType(entity.FieldTypeJSON),
		).
		WithField(entity.NewField().
			WithName("created_at").
			WithDataType(entity.FieldTypeVarChar).
			WithTypeParams(entity.TypeParamMaxLength, "64"),
		)

	// 创建集合
	err = c._client.CreateCollection(ctx, milvusclient.NewCreateCollectionOption(collectionName, schema))
	if err != nil {
		span.RecordError(err)
		return err
	}

	// 创建向量索引
	idx := index.NewHNSWIndex(entity.L2, 16, 256)
	_, err = c._client.CreateIndex(ctx, milvusclient.NewCreateIndexOption(collectionName, "vector", idx))
	if err != nil {
		span.RecordError(err)
		return err
	}

	// 加载集合到内存
	_, err = c._client.LoadCollection(ctx, milvusclient.NewLoadCollectionOption(collectionName))
	if err != nil {
		span.RecordError(err)
		return err
	}

	span.SetAttributes(attribute.String("collectionName", collectionName), attribute.Int("dimension", dimension))
	return nil
}

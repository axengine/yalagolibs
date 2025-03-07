package mgocli

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// UpdateOption 定义更新选项
type UpdateOption struct {
	collection string
	filter     bson.D
	update     bson.D
}

// UpdateBuilder 更新构建器
type UpdateBuilder struct {
	opt *UpdateOption
}

// NewUpdateBuilder 创建更新构建器
func NewUpdateBuilder(collection string) *UpdateBuilder {
	return &UpdateBuilder{
		opt: &UpdateOption{
			collection: collection,
			filter:     bson.D{},
			update:     bson.D{{Key: "$set", Value: bson.D{}}},
		},
	}
}

// Filter 添加过滤条件
func (b *UpdateBuilder) Filter(key string, value interface{}) *UpdateBuilder {
	b.opt.filter = append(b.opt.filter, bson.E{Key: key, Value: value})
	return b
}

// Set 添加更新字段
func (b *UpdateBuilder) Set(key string, value interface{}) *UpdateBuilder {
	if len(b.opt.update) == 0 {
		b.opt.update = bson.D{{Key: "$set", Value: bson.D{}}}
	}
	b.opt.update[0].Value = append(b.opt.update[0].Value.(bson.D), bson.E{Key: key, Value: value})
	return b
}

// Push 添加数组推送操作
func (b *UpdateBuilder) Push(key string, value interface{}) *UpdateBuilder {
	b.opt.update = append(b.opt.update, bson.E{Key: "$push", Value: bson.D{{Key: key, Value: value}}})
	return b
}

// Build 构建更新选项
func (b *UpdateBuilder) Build() *UpdateOption {
	// 自动添加更新时间
	b.Set("updatedAt", time.Now())
	return b.opt
}

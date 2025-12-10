package mgocli

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// UpdateOption represents update operation options
type UpdateOption struct {
	collection string
	filter     bson.D
	update     bson.D
}

// UpdateBuilder represents a MongoDB update operation builder
type UpdateBuilder struct {
	opt *UpdateOption
}

// NewUpdateBuilder creates a new update builder instance
func NewUpdateBuilder(collection string) *UpdateBuilder {
	return &UpdateBuilder{
		opt: &UpdateOption{
			collection: collection,
			filter:     bson.D{},
			update:     bson.D{{Key: "$set", Value: bson.D{}}},
		},
	}
}

// Filter adds a filter condition to the update operation
func (b *UpdateBuilder) Filter(key string, value interface{}) *UpdateBuilder {
	b.opt.filter = append(b.opt.filter, bson.E{Key: key, Value: value})
	return b
}

// Set adds a field update operation
func (b *UpdateBuilder) Set(key string, value interface{}) *UpdateBuilder {
	if len(b.opt.update) == 0 {
		b.opt.update = bson.D{{Key: "$set", Value: bson.D{}}}
	}
	b.opt.update[0].Value = append(b.opt.update[0].Value.(bson.D), bson.E{Key: key, Value: value})
	return b
}

// Push adds an array push operation
func (b *UpdateBuilder) Push(key string, value interface{}) *UpdateBuilder {
	b.opt.update = append(b.opt.update, bson.E{Key: "$push", Value: bson.D{{Key: key, Value: value}}})
	return b
}

// Build constructs and returns the update options
// Automatically adds updatedAt timestamp
func (b *UpdateBuilder) Build() *UpdateOption {
	b.Set("updatedAt", time.Now())
	return b.opt
}

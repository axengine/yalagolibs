package mgocli

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// QueryBuilder represents a MongoDB query builder
type QueryBuilder struct {
	collection string
	filter     bson.D
	opts       *options.FindOptionsBuilder
}

// NewQueryBuilder creates a new query builder instance
func NewQueryBuilder(collection string) *QueryBuilder {
	return &QueryBuilder{
		collection: collection,
		filter:     bson.D{},
		opts:       options.Find(),
	}
}

// Filter adds a filter condition to the query
func (b *QueryBuilder) Filter(key string, value interface{}) *QueryBuilder {
	b.filter = append(b.filter, bson.E{Key: key, Value: value})
	return b
}

// Sort adds a sort condition to the query
func (b *QueryBuilder) Sort(key string, ascending bool) *QueryBuilder {
	value := 1
	if !ascending {
		value = -1
	}
	b.opts.SetSort(bson.D{{Key: key, Value: value}})
	return b
}

// Page sets pagination parameters for the query
func (b *QueryBuilder) Page(page, pageSize int64) *QueryBuilder {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	b.opts.SetSkip((page - 1) * pageSize)
	b.opts.SetLimit(pageSize)
	return b
}

// Build constructs and returns the query parameters
func (b *QueryBuilder) Build() (string, bson.D, *options.FindOptionsBuilder) {
	return b.collection, b.filter, b.opts
}

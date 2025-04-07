package mgocli

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type QueryOneBuilder struct {
	collection string
	filter     bson.D
	opts       *options.FindOneOptionsBuilder
}

func NewQueryOneBuilder(collection string) *QueryOneBuilder {
	return &QueryOneBuilder{
		collection: collection,
		filter:     bson.D{},
		opts:       options.FindOne(),
	}
}

func (b *QueryOneBuilder) Where(filter bson.D) *QueryOneBuilder {
	b.filter = filter
	return b
}

func (b *QueryOneBuilder) Sort(sort interface{}) *QueryOneBuilder {
	b.opts.SetSort(sort)
	return b
}

func (b *QueryOneBuilder) Skip(skip int64) *QueryOneBuilder {
	b.opts.SetSkip(skip)
	return b
}

func (b *QueryOneBuilder) Projection(projection interface{}) *QueryOneBuilder {
	b.opts.SetProjection(projection)
	return b
}

func (b *QueryOneBuilder) Build() (string, bson.D, *options.FindOneOptionsBuilder) {
	return b.collection, b.filter, b.opts
}

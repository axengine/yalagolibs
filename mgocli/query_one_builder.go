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

func (b *QueryOneBuilder) Filter(filter bson.D) *QueryOneBuilder {
	b.filter = filter
	return b
}

func (b *QueryOneBuilder) FilterE(key string, value interface{}) *QueryOneBuilder {
	b.filter = append(b.filter, bson.E{Key: key, Value: value})
	return b
}

func (b *QueryOneBuilder) FilterElements(elements ...bson.E) *QueryOneBuilder {
	b.filter = append(b.filter, elements...)
	return b
}

// Sort adds a sort condition to the query
func (b *QueryOneBuilder) Sort(key string, ascending bool) *QueryOneBuilder {
	value := 1
	if !ascending {
		value = -1
	}
	b.opts.SetSort(bson.D{{Key: key, Value: value}})
	return b
}

// SortE adds sort conditions using bson.E
func (b *QueryOneBuilder) SortE(elements ...bson.E) *QueryOneBuilder {
	b.opts.SetSort(bson.D(elements))
	return b
}

func (b *QueryOneBuilder) SetSkip(skip int64) *QueryOneBuilder {
	b.opts.SetSkip(skip)
	return b
}

func (b *QueryOneBuilder) Projection(projection interface{}) *QueryOneBuilder {
	b.opts.SetProjection(projection)
	return b
}

func (b *QueryOneBuilder) ProjectionE(key string, value interface{}) *QueryOneBuilder {
	return b.Projection(bson.D{{Key: key, Value: value}})
}

func (b *QueryOneBuilder) Build() (string, bson.D, *options.FindOneOptionsBuilder) {
	return b.collection, b.filter, b.opts
}

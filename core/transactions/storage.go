package transactions

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrNotFound = errors.New("not found")

type MongoStorage struct {
	db *mongo.Database
}

func NewMongoStorage(db *mongo.Database) *MongoStorage {
	return &MongoStorage{
		db: db,
	}
}

func (s *MongoStorage) transactions() *mongo.Collection {
	return s.db.Collection("transactions")
}

func convertError(err error) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ErrNotFound
	}

	if errors.Is(err, mongo.ErrNilDocument) {
		return ErrNotFound
	}

	return err
}

func (s *MongoStorage) BulkInsertTXs(
	ctx context.Context,
	data []any,
) error {
	_, err := s.transactions().InsertMany(ctx, data)
	if err != nil {
		return convertError(err)
	}
	return nil
}

func (s *MongoStorage) FindData(ctx context.Context, filter TXFilter) (TXs []TX, err error) {
	// sort data by blockNumber desc
	op := options.Find()
	op.Sort = bson.M{"blockNumber": -1}

	// pagination
	op.SetSkip(filter.Page * filter.PageSize)
	op.SetLimit(filter.PageSize)

	// main filter
	filterBSON := bson.M{}
	if filter.Hash != "" {
		filterBSON["hash"] = filter.Hash
	}
	if filter.From != "" {
		filterBSON["from"] = filter.From
	}
	if filter.To != "" {
		filterBSON["to"] = filter.To
	}
	if filter.BlockNum > 0 {
		filterBSON["blockNumber"] = filter.BlockNum
	}

	dateFilter := bson.M{}
	if !filter.DateFrom.IsZero() {
		dateFilter["$gte"] = filter.DateFrom
	}
	if !filter.DateTo.IsZero() {
		dateFilter["$lte"] = filter.DateTo
	}
	if len(dateFilter) != 0 {
		filterBSON["timestamp"] = dateFilter
	}

	cur, err := s.transactions().Find(ctx, filterBSON, op)
	if err != nil {
		return nil, convertError(err)
	}
	err = cur.All(ctx, &TXs)
	if err != nil {
		return nil, err
	}
	return
}

func (s *MongoStorage) FindLastBlockNumAndConfirm(ctx context.Context) (num int64, confirm int64, err error) {
	op := options.FindOne().SetProjection(bson.M{"blockNumber": 1, "confirmations": 1})
	op.Sort = bson.M{"blockNumber": -1}

	var tx TX
	err = s.transactions().FindOne(ctx, bson.M{}, op).Decode(&tx)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, 0, nil
		}
		return 0, 0, err
	}

	return tx.BlockNumber, tx.Confirmations, nil
}

func (s *MongoStorage) IncAllTXsConf(ctx context.Context) (err error) {
	update := bson.D{{"$inc", bson.D{{"confirmations", 1}}}}
	_, err = s.transactions().UpdateMany(ctx, bson.D{}, update)
	return
}
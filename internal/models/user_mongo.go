package models

import (
	"context"

	"github.com/MitP1997/golang-user-management/internal/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func (u *User) Insert(ctx context.Context) (err *errors.Error) {
	now := timestamppb.Now()
	u.CreatedAt = now
	u.UpdatedAt = now

	e := validator.Struct(u)
	if e != nil {
		return errors.ValidationError(e)
	}

	res, e := userCollection.InsertOne(ctx, u)
	if e != nil {
		return getErrorToReturn(e, "user")
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		u.Id = oid.Hex()
	}
	return nil
}

func (u *User) FindOne(ctx context.Context, filter bson.M) (err *errors.Error) {
	filter, err = updateFilterIdToObjectIdIfExists(filter)
	if err != nil {
		return err
	}
	e := userCollection.FindOne(ctx, filter).Decode(u)
	if e != nil {
		return getErrorToReturn(e, "user")
	}
	return
}

func (u *User) Update(ctx context.Context, filter bson.M, set bson.M) (err *errors.Error) {
	filter, err = updateFilterIdToObjectIdIfExists(filter)
	if err != nil {
		return err
	}
	set = bson.M{"$set": set}
	e := userCollection.FindOneAndUpdate(ctx, filter, set, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(u)
	if e != nil {
		return getErrorToReturn(e, "user")
	}
	return
}

func updateFilterIdToObjectIdIfExists(filter bson.M) (bson.M, *errors.Error) {
	if _, ok := filter["_id"]; ok {
		if _, ok = filter["_id"].(primitive.ObjectID); !ok {
			oid, e := primitive.ObjectIDFromHex(filter["_id"].(string))
			if e != nil {
				return filter, errors.IncorrectIdFormatError(e)
			}
			filter["_id"] = oid
		}
	}
	return filter, nil
}

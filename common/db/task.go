package db

import (
	"context"
	"time"

	"github.com/EutropicAI/FinalRip/module/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CheckTaskExist checks if a task exists in the database
func CheckTaskExist(videoKey string) bool {
	coll := db.DB.Collection(TASK_COLLECTION)
	count, _ := coll.CountDocuments(context.TODO(), Task{Key: videoKey})
	return count > 0
}

// CheckTaskStart checks if a task has started
func CheckTaskStart(videoKey string) bool {
	task, err := GetTask(videoKey)
	if err != nil {
		return false
	}
	return task.EncodeParam != ""
}

// CheckTaskComplete checks if a task has completed
func CheckTaskComplete(videoKey string) bool {
	task, err := GetTask(videoKey)
	if err != nil {
		return false
	}
	return task.EncodeKey != ""
}

// InsertTask inserts a new uncompleted task into the database
func InsertTask(videoKey string) error {
	coll := db.DB.Collection(TASK_COLLECTION)
	_, err := coll.InsertOne(context.TODO(), Task{
		Key:       videoKey,
		CreatedAt: time.Now(),
	})
	return err
}

// UpdateTask updates an uncompleted task in the database
func UpdateTask(filter Task, update Task) error {
	coll := db.DB.Collection(TASK_COLLECTION)

	up := bson.D{{"$set", update}} //nolint:govet

	_, err := coll.UpdateOne(context.TODO(), filter, up)
	if err != nil {
		return err
	}
	return nil
}

// GetTask gets a completed Task from the database
func GetTask(videoKey string) (Task, error) {
	coll := db.DB.Collection(TASK_COLLECTION)
	var task Task
	err := coll.FindOne(context.TODO(), Task{Key: videoKey}).Decode(&task)

	return task, err
}

// DeleteTask deletes a task from the database
func DeleteTask(videoKey string) error {
	coll := db.DB.Collection(TASK_COLLECTION)
	_, err := coll.DeleteOne(context.TODO(), Task{Key: videoKey})
	return err
}

// ListTask list all tasks in the database
func ListTask() ([]Task, error) {
	coll := db.DB.Collection(TASK_COLLECTION)
	cursor, err := coll.Find(context.TODO(), bson.D{},
		options.Find().SetSort(bson.D{{"created_at", -1}})) //nolint:govet

	if err != nil {
		return nil, err
	}

	tasks := make([]Task, 0)
	if err = cursor.All(context.Background(), &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// UnsetTaskEncodeKey 清除任务的 EncodeKey
func UnsetTaskEncodeKey(videoKey string) error {
	coll := db.DB.Collection(TASK_COLLECTION)
	_, err := coll.UpdateOne(context.TODO(), Task{Key: videoKey},
		bson.D{{"$unset", bson.D{{"encode_key", ""}}}}) //nolint: govet
	return err
}

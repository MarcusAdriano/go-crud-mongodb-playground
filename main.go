package main

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Trainer struct {
	Name string
	Age  int
	City string
}

type TrainerRepository interface {
	Save(trainer *Trainer) error
	FindByName(name string) (*Trainer, error)
	FindAll() ([]*Trainer, error)
	DeleteByName(name string) error
}

// TrainerRepository Mongodb Implementation
type TrainerRepositoryMongodb struct {
	mongoCli   *mongo.Client
	db         *mongo.Database
	collection *mongo.Collection
}

func NewTrainerRepository(cli *mongo.Client) (*TrainerRepositoryMongodb, error) {

	db := cli.Database("dbtrainers")
	collection := db.Collection("trainers")

	return &TrainerRepositoryMongodb{mongoCli: cli, db: db, collection: collection}, nil
}

func (repository *TrainerRepositoryMongodb) Save(trainer *Trainer) error {
	insertResults, err := repository.collection.InsertOne(context.TODO(), trainer)
	if err != nil {
		return err
	}

	log.Println("Inserted:", trainer.Name, insertResults)
	return nil
}

func (repository *TrainerRepositoryMongodb) FindAll() ([]*Trainer, error) {
	var results []*Trainer

	findOptions := options.Find()
	findOptions.SetLimit(100_000)

	cursor, err := repository.collection.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		return nil, err
	}

	for cursor.Next(context.TODO()) {
		var trainer Trainer
		err := cursor.Decode(&trainer)
		if err != nil {
			return nil, err
		}

		results = append(results, &trainer)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	cursor.Close(context.TODO())
	return results, nil
}

func (repository *TrainerRepositoryMongodb) FindByName(name string) (*Trainer, error) {
	findOptions := options.FindOne()

	result := repository.collection.FindOne(context.TODO(), bson.M{"name": name}, findOptions)

	var trainer Trainer
	err := result.Decode(&trainer)
	return &trainer, err
}

func (repository *TrainerRepositoryMongodb) DeleteByName(name string) error {
	deletedResult, err := repository.collection.DeleteOne(context.TODO(),
		bson.M{"name": name})

	log.Println("Deleted:", deletedResult)
	return err
}

func errorHandler(err error) {
	if err != nil {
		log.Println("Error:", err)
	}
}

// Bson documents
// D: A BSON document. This type should be used in situations where order matters, such as MongoDB commands.
// M: An unordered map. It is the same as D, except it does not preserve order.
// A: A BSON array.
// E: A single element inside a D.
func main() {
	cliOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	cli, err := mongo.Connect(context.TODO(), cliOptions)
	errorHandler(err)

	err = cli.Ping(context.TODO(), nil)
	errorHandler(err)

	db := cli.Database("dbtrainers")
	defer db.Drop(context.TODO())

	err = db.CreateCollection(context.TODO(), "trainers")
	errorHandler(err)

	var trainersRepo TrainerRepository
	trainersRepo, err = NewTrainerRepository(cli)
	errorHandler(err)

	trainersRepo.Save(&Trainer{Name: "Marcus", Age: 25, City: "Nuporanga-SP"})
	trainersRepo.Save(&Trainer{Name: "Leticia Presoto", Age: 25, City: "Orlandia-SP"})
	trainersRepo.Save(&Trainer{Name: "Magali", Age: 2, City: "Uberlandia-MG"})
	trainersRepo.Save(&Trainer{Name: "Cacau", Age: 1, City: "Uberlandia-MG"})

	trainers, err := trainersRepo.FindAll()
	errorHandler(err)

	for _, trainer := range trainers {
		log.Println(trainer)
	}

	found, err := trainersRepo.FindByName("Marcus")
	errorHandler(err)
	log.Println("Found:", found)

	trainersRepo.DeleteByName("Marcus")
}

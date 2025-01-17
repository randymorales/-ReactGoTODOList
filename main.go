package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Task struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Completed   bool               `json:"completed"`
	Description string             `json:"description"`
}

var collection *mongo.Collection

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading the .env file: ", err)
	}

	mongodbURI := os.Getenv("MONGODB_URI")
	clientOptions := options.Client().ApplyURI(mongodbURI)

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}

	defer client.Disconnect(context.Background())

	if err = client.Ping(context.Background(), nil); err != nil {
		log.Fatal("Error connecting to the client: ", err)
	}

	fmt.Println("Connected to MongoDB!")
	collection = client.Database("todolist_db").Collection("tasks")

	app := fiber.New()
	app.Get("/api/tasks", getTasks)
	app.Post("/api/create", createTask)
	app.Patch("/api/update/:id", updateTask)
	app.Delete("/api/delete/:id", deleteTask)

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	log.Fatal(app.Listen("0.0.0.0:" + port))
}

func getTasks(c *fiber.Ctx) error {
	var tasks []Task

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var task Task
		if err := cursor.Decode(&task); err != nil {
			return err
		}

		tasks = append(tasks, task)
	}

	return c.JSON(tasks)
}

func createTask(c *fiber.Ctx) error {
	task := new(Task)

	if err := c.BodyParser(task); err != nil {
		return err
	}

	if task.Description == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Task description is required"})
	}

	result, err := collection.InsertOne(context.Background(), task)
	if err != nil {
		return err
	}

	task.ID = result.InsertedID.(primitive.ObjectID)
	return c.Status(201).JSON(task)
}

func updateTask(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"completed": true}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"sucess": true})
}

func deleteTask(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Task not found"})
	}

	filter := bson.M{"_id": objectID}
	_, err = collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"sucess": true})
}

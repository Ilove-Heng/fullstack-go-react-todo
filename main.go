package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Todo represents a todo item
type Todo struct {
	ID        primitive.ObjectID       `json:"_id,omitempty" bson:"_id,omitempty"`
	Completed bool      `json:"completed"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

var collection *mongo.Collection

func main() {
	// load .env if not in production
	if os.Getenv("ENV") != "production" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file:", err)
		}
	}

	// connect to mongodb
	MONGODB_URI := os.Getenv("MONGODB_URI")
	clientOptions := options.Client().ApplyURI(MONGODB_URI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// disconnect when main returns
	defer client.Disconnect(context.Background())

	// check connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to mongodb...")

	// set collection
	collection = client.Database("golang_db").Collection("todos")

	// create fiber app
	app := fiber.New()

	// add CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders: "Origin,Content-Type,Accept",
	}))

	// routes
	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", createTodo)
	app.Patch("/api/todos/:id", updateTodo)
	app.Delete("/api/todos/:id", deleteTodo)

	// serve static files in production
	if os.Getenv("ENV") == "production" {
		app.Static("/", "./client/dist")
	}

	// start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	log.Fatal(app.Listen("0.0.0.0:" + port))
}

func getTodos(c *fiber.Ctx) error {
	var todos []Todo

	// find all todos
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	// defer cursor close
	defer cursor.Close(context.Background())

	// loop through cursor and append to todos
	for cursor.Next(context.Background()) {
		var todo Todo
		if err := cursor.Decode(&todo); err != nil {
			return err
		}
		todos = append(todos, todo)
	}

	// return todos as JSON
	return c.JSON(todos)
}

func createTodo(c *fiber.Ctx) error {
	todo := new(Todo)

	// parse body
	if err := c.BodyParser(todo); err != nil {
		return err
	}

	// check body is required
	if todo.Body == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Body is required",
		})
	}

	// insert todo
	insertResult, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		return err
	}

	// set ID
	todo.ID = insertResult.InsertedID.(primitive.ObjectID)

	// return todo as JSON
	return c.Status(201).JSON(todo)
}

func updateTodo(c *fiber.Ctx) error {
	// get ID from params
	id := c.Params("id")

	// convert ID to primitive.ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(422).JSON(fiber.Map{
			"error": "Invalid todo ID format",
			"details": err.Error(),
		})
	}

	// create update document
	update := bson.M{
		"$set": bson.M{
			"completed":  true,
			"updated_at": time.Now(),
		},
	}

	// find and update todo
	filter := bson.M{"_id": objectID}
	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update todo",
			"details": err.Error(),
		})
	}

	// check if todo was found
	if result.MatchedCount == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "Todo not found",
		})
	}

	// get updated todo
	var updatedTodo Todo
	err = collection.FindOne(context.Background(), filter).Decode(&updatedTodo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch updated todo",
			"details": err.Error(),
		})
	}

	// return updated todo as JSON
	return c.Status(200).JSON(updatedTodo)
}

func deleteTodo(c *fiber.Ctx) error {
	id := c.Params("id")

	// convert ID to primitive.ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// create filter
	filter := bson.M{"_id": objectID}

	// delete todo
	result, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}

	// check if todo was found
	if result.DeletedCount == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "Todo not found",
		})
	}

	// return success message
	return c.Status(200).JSON(fiber.Map{
		"message": "Todo deleted successfully",
	})
}


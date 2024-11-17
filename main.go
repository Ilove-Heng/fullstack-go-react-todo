package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID        primitive.ObjectID       `json:"id,omitempty" bson:"_id,omitempty"`
	Completed bool      `json:"completed"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

var collection *mongo.Collection

func main() {

	// load .env
	err := godotenv.Load(".env");

	// env not found 
	if err != nil {
		log.Fatal("Error loading .env file:", err);
	}

	MONGODB_URI := os.Getenv("MONGODB_URI");
	clientOptions := options.Client().ApplyURI(MONGODB_URI);

	client, err := mongo.Connect(context.Background(), clientOptions);

	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to mongodb...")

	collection = client.Database("golang_db").Collection("todos");

	// load fiber
	app := fiber.New()

	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", createTodo)
	app.Patch("/api/todos/:id", updateTodo)
	app.Delete("/api/todos/:id", deleteTodo)

	// load PORT
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	log.Fatal(app.Listen("0.0.0.0:" + port))
}

func getTodos(c *fiber.Ctx) error {
	var todos []Todo

	cursor, err := collection.Find(context.Background(), bson.M{})

	if err != nil {
		return err;
	}

	// optimization - defer cursor.Close(context.Background())
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var todo Todo;
		if err := cursor.Decode(&todo); err != nil {
			return err;
		}
		todos = append(todos, todo)
	}

	return c.JSON(todos)
}

func createTodo(c *fiber.Ctx) error {
	todo := new(Todo)

	if err := c.BodyParser(todo); err != nil {
		return err;
	}

		// check body is required
		if todo.Body == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Body is required",
			})
		}

	insertResult, err := collection.InsertOne(context.Background(), todo)

	if err != nil {
		return err;
	}

	todo.ID = insertResult.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(todo)
}

func updateTodo(c *fiber.Ctx) error {
    // Get the ID from parameters
    id := c.Params("id")
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return c.Status(422).JSON(fiber.Map{
            "error": "Invalid todo ID format",
            "details": err.Error(),
        })
    }

    // Create a struct to hold the update data
    type UpdateTodoInput struct {
        Body      *string `json:"body"`      // Using pointer to distinguish between empty and missing fields
        Completed *bool   `json:"completed"` // Using pointer to distinguish between false and missing fields
    }

    // Parse the request body
    var input UpdateTodoInput
    if err := c.BodyParser(&input); err != nil {
        return c.Status(422).JSON(fiber.Map{
            "error": "Failed to parse request body",
            "details": err.Error(),
        })
    }

    // Validate that at least one field is provided
    if input.Body == nil && input.Completed == nil {
        return c.Status(422).JSON(fiber.Map{
            "error": "At least one field (body or completed) must be provided for update",
        })
    }

    // Validate body length if it's provided
    if input.Body != nil && len(*input.Body) < 1 {
        return c.Status(422).JSON(fiber.Map{
            "error": "Body cannot be empty",
        })
    }

    // Create update document
    update := bson.M{"$set": bson.M{}}
    updateDoc := update["$set"].(bson.M)

    // Only include fields that were provided in the request
    if input.Body != nil {
        updateDoc["body"] = *input.Body
    }
    if input.Completed != nil {
        updateDoc["completed"] = *input.Completed
    }

    // Add last updated timestamp
    updateDoc["updated_at"] = time.Now()

    // Find and update the document
    filter := bson.M{"_id": objectID}
    result, err := collection.UpdateOne(context.Background(), filter, update)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to update todo",
            "details": err.Error(),
        })
    }

    if result.MatchedCount == 0 {
        return c.Status(404).JSON(fiber.Map{
            "error": "Todo not found",
        })
    }

    // Get the updated todo
    var updatedTodo Todo
    err = collection.FindOne(context.Background(), filter).Decode(&updatedTodo)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to fetch updated todo",
            "details": err.Error(),
        })
    }

    return c.Status(200).JSON(updatedTodo)
}

func deleteTodo(c *fiber.Ctx) error {
	id := c.Params("id");

	objectID, err := primitive.ObjectIDFromHex(id);

	if err != nil {
		return err;
	}

	filter := bson.M{"_id": objectID}

	result, err := collection.DeleteOne(context.Background(), filter)

	if err != nil {
		return err;
	}

	if result.DeletedCount == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "Todo not found",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "Todo deleted successfully",
	})
}


package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

type Todo struct {
	ID        int       `json:"id"`
	Completed bool      `json:"completed"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {

	// load fiber
	app := fiber.New()

	// load env
	err := godotenv.Load(".env");

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	PORT := os.Getenv("PORT");

	todos := []Todo{}

	app.Get("/api/todos", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"success":    true,
			"message":    "Get Successfully",
			"status_code": 200,
			"data":        todos,
		})
	})

	// Create a todo
	app.Post("/api/todos", func(c *fiber.Ctx) error {
		todo := &Todo{}

		// check body parser
		if err := c.BodyParser(todo); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error":   "Invalid request format",
				"details": err.Error(),
			})
		}

		// check body is required
		if todo.Body == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Body is required",
			})
		}

		todo.ID = len(todos) + 1
		todos = append(todos, *todo)

		fmt.Println("todos", todo)

		return c.Status(201).JSON(todo)
	})

	// Update a todo
	app.Patch("/api/todos/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		for i, todo := range todos {
			if fmt.Sprint(todo.ID) == id {
				todos[i].Completed = true
				return c.Status(200).JSON(todos[i])
			}
		}

		return c.Status(404).JSON(fiber.Map{
			"error": "Todo not found",
		})
	})

	// Delete a todo
	app.Delete("/api/todos/:id", func(c *fiber.Ctx) error {
		var id string = c.Params("id")

		for i, todo := range todos {
			if fmt.Sprint(todo.ID) == id {
				todos = append(todos[:i], todos[i+1:]...)
				return c.Status(200).JSON(fiber.Map{
					"success": true,
				})
			}
		}
		return c.Status(404).JSON(fiber.Map{
			"error": "Todo not found",
		})
	})

	log.Fatal(app.Listen(":" + PORT))
}


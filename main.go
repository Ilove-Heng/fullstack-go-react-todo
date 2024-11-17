package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Todo struct {
	ID        int       `json:"id"`
	Completed bool      `json:"completed"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	fmt.Println("Hello, World haha!")

	app := fiber.New()

	todos := []Todo{}

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"message": "Hello, World!",
		})
	})

	app.Post("/api/todos", func(c *fiber.Ctx) error {
		todo := &Todo{} // memory address

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

	log.Fatal(app.Listen(":4000"))
}

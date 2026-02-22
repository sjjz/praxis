package api

import (
	"praxis/app/lib"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
)

func (s *Server) Router() *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Use(cors.New())
	app.Get("/v1/health", s.health)

	v1 := app.Group("/v1", s.devUserMiddleware)
	v1.Post("/mood-checkins", s.createMoodCheckin)
	v1.Get("/mood-checkins", s.listMoodCheckins)
	v1.Patch("/mood-checkins/:id", s.updateMoodCheckin)
	v1.Delete("/mood-checkins/:id", s.deleteMoodCheckin)

	v1.Post("/nutrition-entries", s.createNutritionEntry)
	v1.Get("/nutrition-entries", s.listNutritionEntries)
	v1.Patch("/nutrition-entries/:id", s.updateNutritionEntry)
	v1.Delete("/nutrition-entries/:id", s.deleteNutritionEntry)

	v1.Get("/daily-summaries", s.dailySummaries)
	v1.Get("/trends", s.trends)

	return app
}

func (s *Server) devUserMiddleware(c *fiber.Ctx) error {
	c.Locals("userID", s.cfg.DevUserID)
	return c.Next()
}

func userIDFromCtx(c *fiber.Ctx) uuid.UUID {
	v := c.Locals("userID")
	if v == nil {
		return uuid.Nil
	}
	id, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return id
}

func respondErr(c *fiber.Ctx, status int, msg string, details any) error {
	return c.Status(status).JSON(lib.APIError{Error: msg, Details: details})
}

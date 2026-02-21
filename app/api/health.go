package api

import "github.com/gofiber/fiber/v2"

func (s *Server) health(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
}

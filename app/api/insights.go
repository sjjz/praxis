package api

import (
	"time"

	"praxis/app/lib"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) dailySummaries(c *fiber.Ctx) error {
	fromQ := c.Query("from")
	toQ := c.Query("to")
	if fromQ == "" || toQ == "" {
		return respondErr(c, fiber.StatusBadRequest, "from and to are required (YYYY-MM-DD)", nil)
	}
	from, err := time.Parse("2006-01-02", fromQ)
	if err != nil {
		return respondErr(c, fiber.StatusBadRequest, "from must be YYYY-MM-DD", nil)
	}
	to, err := time.Parse("2006-01-02", toQ)
	if err != nil {
		return respondErr(c, fiber.StatusBadRequest, "to must be YYYY-MM-DD", nil)
	}
	if to.Before(from) {
		return respondErr(c, fiber.StatusBadRequest, "to must be on or after from", nil)
	}

	userID := userIDFromCtx(c)
	tz, err := s.store.UserTimezone(c.Context(), userID)
	if err != nil {
		return respondErr(c, fiber.StatusInternalServerError, "failed to load timezone", nil)
	}
	summaries, err := s.store.DailySummaries(c.Context(), userID, from, to, tz)
	if err != nil {
		return respondErr(c, fiber.StatusInternalServerError, "failed to compute daily summaries", nil)
	}
	return c.JSON(fiber.Map{"items": summaries})
}

func (s *Server) trends(c *fiber.Ctx) error {
	window := c.Query("window")
	if window != "7d" && window != "30d" {
		return respondErr(c, fiber.StatusBadRequest, "window must be 7d or 30d", nil)
	}
	days := 7
	if window == "30d" {
		days = 30
	}
	userID := userIDFromCtx(c)
	tz, err := s.store.UserTimezone(c.Context(), userID)
	if err != nil {
		return respondErr(c, fiber.StatusInternalServerError, "failed to load timezone", nil)
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	nowLocal := time.Now().In(loc)
	to := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 0, 0, 0, 0, loc)
	from := to.AddDate(0, 0, -(days - 1))
	summaries, err := s.store.DailySummaries(c.Context(), userID, from, to, tz)
	if err != nil {
		return respondErr(c, fiber.StatusInternalServerError, "failed to compute trends", nil)
	}
	return c.JSON(lib.TrendsResponse{Window: window, Points: summaries})
}

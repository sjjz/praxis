package api

import (
	"errors"
	"time"

	"praxis/app/lib"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Server) createMoodCheckin(c *fiber.Ctx) error {
	var req lib.CreateMoodCheckinRequest
	if err := c.BodyParser(&req); err != nil {
		return respondErr(c, fiber.StatusBadRequest, "invalid JSON body", nil)
	}
	if !lib.IsValidMoodType(req.MoodType) {
		return respondErr(c, fiber.StatusUnprocessableEntity, "invalid mood_type", nil)
	}
	if req.Quality < 1 || req.Quality > 5 {
		return respondErr(c, fiber.StatusUnprocessableEntity, "quality must be between 1 and 5", nil)
	}
	ts, err := lib.ParseTimestamp(req.Timestamp)
	if err != nil {
		return respondErr(c, fiber.StatusUnprocessableEntity, err.Error(), nil)
	}
	req.Timestamp = ts.Format(time.RFC3339Nano)
	item, err := s.store.CreateMoodCheckin(c.Context(), userIDFromCtx(c), req)
	if err != nil {
		return respondErr(c, fiber.StatusInternalServerError, "failed to create mood checkin", nil)
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (s *Server) listMoodCheckins(c *fiber.Ctx) error {
	limit, err := lib.ParseLimit(c.Query("limit"), 50)
	if err != nil {
		return respondErr(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	cursor, err := lib.DecodeCursor(c.Query("cursor"))
	if err != nil {
		return respondErr(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	var from, to *time.Time
	if q := c.Query("from"); q != "" {
		ts, parseErr := lib.ParseTimestamp(q)
		if parseErr != nil {
			return respondErr(c, fiber.StatusBadRequest, parseErr.Error(), nil)
		}
		from = &ts
	}
	if q := c.Query("to"); q != "" {
		ts, parseErr := lib.ParseTimestamp(q)
		if parseErr != nil {
			return respondErr(c, fiber.StatusBadRequest, parseErr.Error(), nil)
		}
		to = &ts
	}
	var moodType *lib.MoodType
	if q := c.Query("mood_type"); q != "" {
		mt := lib.MoodType(q)
		if !lib.IsValidMoodType(mt) {
			return respondErr(c, fiber.StatusBadRequest, "invalid mood_type", nil)
		}
		moodType = &mt
	}

	items, next, err := s.store.ListMoodCheckins(c.Context(), userIDFromCtx(c), lib.ListMoodFilters{
		From: from, To: to, MoodType: moodType, Cursor: cursor, Limit: limit,
	})
	if err != nil {
		return respondErr(c, fiber.StatusInternalServerError, "failed to list mood checkins", nil)
	}
	return c.JSON(fiber.Map{"items": items, "next_cursor": next})
}

func (s *Server) updateMoodCheckin(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respondErr(c, fiber.StatusBadRequest, "invalid id", nil)
	}
	var req lib.UpdateMoodCheckinRequest
	if err := c.BodyParser(&req); err != nil {
		return respondErr(c, fiber.StatusBadRequest, "invalid JSON body", nil)
	}
	if req.MoodType == nil && req.Quality == nil && req.Note == nil && req.Timestamp == nil {
		return respondErr(c, fiber.StatusBadRequest, "at least one field is required", nil)
	}
	if req.MoodType != nil && !lib.IsValidMoodType(*req.MoodType) {
		return respondErr(c, fiber.StatusUnprocessableEntity, "invalid mood_type", nil)
	}
	if req.Quality != nil && (*req.Quality < 1 || *req.Quality > 5) {
		return respondErr(c, fiber.StatusUnprocessableEntity, "quality must be between 1 and 5", nil)
	}
	item, err := s.store.UpdateMoodCheckin(c.Context(), userIDFromCtx(c), id, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return respondErr(c, fiber.StatusNotFound, "mood checkin not found", nil)
		}
		return respondErr(c, fiber.StatusInternalServerError, "failed to update mood checkin", nil)
	}
	return c.JSON(item)
}

func (s *Server) deleteMoodCheckin(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respondErr(c, fiber.StatusBadRequest, "invalid id", nil)
	}
	err = s.store.DeleteMoodCheckin(c.Context(), userIDFromCtx(c), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return respondErr(c, fiber.StatusNotFound, "mood checkin not found", nil)
		}
		return respondErr(c, fiber.StatusInternalServerError, "failed to delete mood checkin", nil)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

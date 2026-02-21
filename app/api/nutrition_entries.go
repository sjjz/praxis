package api

import (
	"errors"
	"strings"
	"time"

	"praxis/app/lib"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Server) createNutritionEntry(c *fiber.Ctx) error {
	var req lib.CreateNutritionEntryRequest
	if err := c.BodyParser(&req); err != nil {
		return respondErr(c, fiber.StatusBadRequest, "invalid JSON body", nil)
	}
	if req.MealTag != nil && !lib.IsValidMealTag(*req.MealTag) {
		return respondErr(c, fiber.StatusUnprocessableEntity, "invalid meal_tag", nil)
	}
	if !lib.AtLeastOneMacro(req) {
		return respondErr(c, fiber.StatusUnprocessableEntity, "at least one macro field is required", nil)
	}
	if err := lib.ValidateNonNegative("calories", req.Calories); err != nil {
		return respondErr(c, fiber.StatusUnprocessableEntity, err.Error(), nil)
	}
	if err := lib.ValidateNonNegative("protein_g", req.ProteinG); err != nil {
		return respondErr(c, fiber.StatusUnprocessableEntity, err.Error(), nil)
	}
	if err := lib.ValidateNonNegative("fiber_g", req.FiberG); err != nil {
		return respondErr(c, fiber.StatusUnprocessableEntity, err.Error(), nil)
	}
	if err := lib.ValidateNonNegative("added_sugar_g", req.AddedSugarG); err != nil {
		return respondErr(c, fiber.StatusUnprocessableEntity, err.Error(), nil)
	}
	if err := lib.ValidateNonNegative("carbs_g", req.CarbsG); err != nil {
		return respondErr(c, fiber.StatusUnprocessableEntity, err.Error(), nil)
	}
	ts, err := lib.ParseTimestamp(req.Timestamp)
	if err != nil {
		return respondErr(c, fiber.StatusUnprocessableEntity, err.Error(), nil)
	}
	req.Timestamp = ts.Format(time.RFC3339Nano)

	item, err := s.store.CreateNutritionEntry(c.Context(), userIDFromCtx(c), req)
	if err != nil {
		return respondErr(c, fiber.StatusInternalServerError, "failed to create nutrition entry", nil)
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (s *Server) listNutritionEntries(c *fiber.Ctx) error {
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
	var mealTag *lib.MealTag
	if q := c.Query("meal_tag"); q != "" {
		mt := lib.MealTag(q)
		if !lib.IsValidMealTag(mt) {
			return respondErr(c, fiber.StatusBadRequest, "invalid meal_tag", nil)
		}
		mealTag = &mt
	}
	items, next, err := s.store.ListNutritionEntries(c.Context(), userIDFromCtx(c), lib.ListNutritionFilters{
		From: from, To: to, MealTag: mealTag, Cursor: cursor, Limit: limit,
	})
	if err != nil {
		return respondErr(c, fiber.StatusInternalServerError, "failed to list nutrition entries", nil)
	}
	return c.JSON(fiber.Map{"items": items, "next_cursor": next})
}

func (s *Server) updateNutritionEntry(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respondErr(c, fiber.StatusBadRequest, "invalid id", nil)
	}
	var req lib.UpdateNutritionEntryRequest
	if err := c.BodyParser(&req); err != nil {
		return respondErr(c, fiber.StatusBadRequest, "invalid JSON body", nil)
	}
	if req.Timestamp == nil && req.MealTag == nil && req.Calories == nil && req.ProteinG == nil &&
		req.FiberG == nil && req.AddedSugarG == nil && req.CarbsG == nil {
		return respondErr(c, fiber.StatusBadRequest, "at least one field is required", nil)
	}
	if req.MealTag != nil && !lib.IsValidMealTag(*req.MealTag) {
		return respondErr(c, fiber.StatusUnprocessableEntity, "invalid meal_tag", nil)
	}
	if err := lib.ValidateNonNegative("calories", req.Calories); err != nil {
		return respondErr(c, fiber.StatusUnprocessableEntity, err.Error(), nil)
	}
	if err := lib.ValidateNonNegative("protein_g", req.ProteinG); err != nil {
		return respondErr(c, fiber.StatusUnprocessableEntity, err.Error(), nil)
	}
	if err := lib.ValidateNonNegative("fiber_g", req.FiberG); err != nil {
		return respondErr(c, fiber.StatusUnprocessableEntity, err.Error(), nil)
	}
	if err := lib.ValidateNonNegative("added_sugar_g", req.AddedSugarG); err != nil {
		return respondErr(c, fiber.StatusUnprocessableEntity, err.Error(), nil)
	}
	if err := lib.ValidateNonNegative("carbs_g", req.CarbsG); err != nil {
		return respondErr(c, fiber.StatusUnprocessableEntity, err.Error(), nil)
	}

	item, err := s.store.UpdateNutritionEntry(c.Context(), userIDFromCtx(c), id, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return respondErr(c, fiber.StatusNotFound, "nutrition entry not found", nil)
		}
		if strings.HasPrefix(err.Error(), "invalid timestamp:") {
			return respondErr(c, fiber.StatusUnprocessableEntity, err.Error(), nil)
		}
		return respondErr(c, fiber.StatusInternalServerError, "failed to update nutrition entry", nil)
	}
	return c.JSON(item)
}

func (s *Server) deleteNutritionEntry(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respondErr(c, fiber.StatusBadRequest, "invalid id", nil)
	}
	err = s.store.DeleteNutritionEntry(c.Context(), userIDFromCtx(c), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return respondErr(c, fiber.StatusNotFound, "nutrition entry not found", nil)
		}
		return respondErr(c, fiber.StatusInternalServerError, "failed to delete nutrition entry", nil)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

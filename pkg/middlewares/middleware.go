package middlewares

import (
	"github.com/gofiber/fiber/v2"
)

func ErrorMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		if err != nil {
			if e, ok := err.(*fiber.Error); ok {
				return c.Status(e.Code).JSON(fiber.Map{
					"message": e.Message,
				})
			}
		}
		return nil
	}
}

func UndefinedRoutesMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		allowedPath := []string{
			"/insertRuleTemplate",
			"/insertRuletoRuleSet",
			"/updateRuleSet",
			"/execInput",
			"/fetchRules",
			"/deleteRuleSet",
		}

		matchedPath := false
		for _, testPath := range allowedPath {
			if c.Path() == "/" {
				return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "application running smoothly"})
			}
			if c.Path() == testPath {
				matchedPath = true
				break
			}
		}

		if !matchedPath {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "resource not found",
			})
		}

		return c.Next()
	}
}

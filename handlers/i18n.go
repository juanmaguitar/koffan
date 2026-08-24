package handlers

import (
	"shopping-list/i18n"

	"github.com/gofiber/fiber/v2"
)

// GetLocales returns list of available languages
func GetLocales(c *fiber.Ctx) error {
	return c.JSON(i18n.AvailableLocales())
}

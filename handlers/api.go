package handlers

import (
	"shopping-list/db"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GetAllData returns all sections with items and stats for offline caching
func GetAllData(c *fiber.Ctx) error {
	sections, err := db.GetAllSections()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch data"})
	}

	stats := db.GetStats()

	return c.JSON(fiber.Map{
		"sections":  sections,
		"stats":     stats,
		"timestamp": time.Now().Unix(),
	})
}

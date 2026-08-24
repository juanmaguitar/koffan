package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"shopping-list/db"
	"shopping-list/i18n"
	"shopping-list/webhook"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Input length limits
const (
	MaxListNameLength    = 100
	MaxIconLength        = 20 // emoji can be multi-byte
	MaxSectionNameLength = 100
	MaxItemNameLength    = 200
	MaxDescriptionLength = 500
)

// defaultSectionName returns the name used for the implicit section that backs
// every list. Sections are hidden in this fork, but items still hang off one
// (items.section_id is NOT NULL), so each list keeps exactly one.
func defaultSectionName() string {
	name := i18n.Get(i18n.GetDefaultLang(), "sections.default")
	if name == "sections.default" {
		name = "General"
	}
	return name
}

// ensureDefaultSection guarantees the list has at least one section to hold
// items, creating the implicit one if needed. Returns the sections to render.
func ensureDefaultSection(listID int64) ([]db.Section, error) {
	sections, err := db.GetSectionsByList(listID)
	if err != nil {
		return nil, err
	}
	if len(sections) > 0 {
		return sections, nil
	}

	if _, err := db.CreateSectionForList(listID, defaultSectionName()); err != nil {
		return nil, err
	}
	return db.GetSectionsByList(listID)
}

// GetListsPage returns the homepage with all lists
func GetListsPage(c *fiber.Ctx) error {
	lists, err := db.GetAllLists()
	if err != nil {
		return sendError(c, 500, "error.fetch_failed")
	}

	templates, _ := db.GetAllTemplates()

	return c.Render("home", fiber.Map{
		"Lists":        lists,
		"Templates":    templates,
		"Translations": i18n.GetAllLocales(),
		"Locales":      i18n.AvailableLocales(),
		"DefaultLang":  i18n.GetDefaultLang(),
	})
}

// GetListView returns a single list with its items
func GetListView(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Redirect("/")
	}

	list, err := db.GetListByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			// List not found - redirect to home
			return c.Redirect("/")
		}
		// Database error - log and show error
		log.Printf("Error fetching list %d: %v", id, err)
		return sendError(c, 500, "error.database_error")
	}

	// Set this list as active
	db.SetActiveList(id)

	// Backfills the implicit section for lists created before this fork.
	sections, err := ensureDefaultSection(id)
	if err != nil {
		return sendError(c, 500, "error.fetch_failed")
	}

	stats := db.GetListStats(id)
	lists, _ := db.GetAllLists()

	return c.Render("list", fiber.Map{
		"List":          list,
		"Lists":         lists,
		"Sections":      sections,
		"Stats":         stats,
		"ShowCompleted": list.ShowCompleted,
		"Translations":  i18n.GetAllLocales(),
		"Locales":       i18n.AvailableLocales(),
		"DefaultLang":   i18n.GetDefaultLang(),
	})
}

// GetLists returns all lists (JSON API)
func GetLists(c *fiber.Ctx) error {
	lists, err := db.GetAllLists()
	if err != nil {
		return sendError(c, 500, "error.fetch_failed")
	}

	// Check if JSON format is requested
	if c.Query("format") == "json" {
		return c.JSON(lists)
	}

	// For HTML, redirect to homepage
	return c.Redirect("/")
}

// CreateList creates a new shopping list
func CreateList(c *fiber.Ctx) error {
	name := c.FormValue("name")
	if name == "" {
		return sendError(c, 400, "error.name_required")
	}
	if len(name) > MaxListNameLength {
		return sendError(c, 400, "error.name_too_long")
	}
	if name == "[HISTORY]" {
		return sendError(c, 400, "common.reserved_name")
	}

	// Check for duplicate name
	exists, err := db.ListNameExists(name, 0)
	if err != nil {
		return sendError(c, 500, "error.check_failed")
	}
	if exists {
		return sendError(c, 409, "list.name_exists")
	}

	icon := c.FormValue("icon")
	if icon == "" {
		icon = "🛒"
	}
	if len(icon) > MaxIconLength {
		return sendError(c, 400, "error.icon_too_long")
	}

	list, err := db.CreateList(name, icon)
	if err != nil {
		return sendError(c, 500, "error.create_failed")
	}

	// A list with no section cannot hold items, and the add bar would have
	// nowhere to post to.
	if _, err := ensureDefaultSection(list.ID); err != nil {
		log.Printf("Error creating default section for list %d: %v", list.ID, err)
	}

	// Broadcast to WebSocket clients
	BroadcastUpdate("list_created", list)

	// Return the new list item partial for HTMX
	return c.Render("partials/list_item", fiber.Map{
		"List": list,
	}, "")
}

// UpdateList updates a list's name and icon
func UpdateList(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return sendError(c, 400, "error.invalid_id")
	}

	name := c.FormValue("name")
	if name == "" {
		return sendError(c, 400, "error.name_required")
	}
	if len(name) > MaxListNameLength {
		return sendError(c, 400, "error.name_too_long")
	}
	if name == "[HISTORY]" {
		return sendError(c, 400, "common.reserved_name")
	}

	// Check for duplicate name (excluding current list)
	exists, err := db.ListNameExists(name, id)
	if err != nil {
		return sendError(c, 500, "error.check_failed")
	}
	if exists {
		return sendError(c, 409, "list.name_exists")
	}

	icon := c.FormValue("icon")
	if len(icon) > MaxIconLength {
		return sendError(c, 400, "error.icon_too_long")
	}

	list, err := db.UpdateList(id, name, icon)
	if err != nil {
		return sendError(c, 500, "error.update_failed")
	}

	// Broadcast to WebSocket clients
	BroadcastUpdate("list_updated", list)

	// Return updated list item partial
	return c.Render("partials/list_item", fiber.Map{
		"List": list,
	}, "")
}

// DeleteList deletes a shopping list
func DeleteList(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return sendError(c, 400, "error.invalid_id")
	}

	preparedWebhooks := PrepareListItemWebhooks(webhook.EventItemDeleted, id)
	err = db.DeleteList(id)
	if err != nil {
		return c.Status(400).SendString(err.Error())
	}

	// Broadcast to WebSocket clients
	BroadcastUpdate("list_deleted", map[string]int64{"id": id})
	NotifyPreparedItemWebhooks(webhook.EventItemDeleted, preparedWebhooks)

	// Return empty string (HTMX will remove the element)
	return c.SendString("")
}

// SetActiveList sets a list as active
func SetActiveList(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return sendError(c, 400, "error.invalid_id")
	}

	err = db.SetActiveList(id)
	if err != nil {
		return sendError(c, 500, "error.check_failed")
	}

	// Broadcast to WebSocket clients
	BroadcastUpdate("list_activated", map[string]int64{"id": id})

	// Check if this is an AJAX request (HTMX or fetch)
	isAjax := c.Get("HX-Request") != "" || c.Get("X-Requested-With") != ""
	if !isAjax {
		return c.Redirect(fmt.Sprintf("/lists/%d", id))
	}

	// Check if this is from the lists management page or main page
	currentURL := c.Get("HX-Current-URL")
	referer := c.Get("Referer")
	isListsPage := strings.Contains(currentURL, "/lists") || strings.Contains(referer, "/lists")

	if !isListsPage {
		c.Set("HX-Redirect", fmt.Sprintf("/lists/%d", id))
		return c.SendString("")
	}

	// Return updated lists for the management page
	return returnAllLists(c)
}

// MoveListUp moves a list up in order
func MoveListUp(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return sendError(c, 400, "error.invalid_id")
	}

	err = db.MoveListUp(id)
	if err != nil {
		return sendError(c, 500, "error.move_failed")
	}

	BroadcastUpdate("lists_reordered", nil)
	return c.SendStatus(200)
}

// MoveListDown moves a list down in order
func MoveListDown(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return sendError(c, 400, "error.invalid_id")
	}

	err = db.MoveListDown(id)
	if err != nil {
		return sendError(c, 500, "error.move_failed")
	}

	BroadcastUpdate("lists_reordered", nil)
	return c.SendStatus(200)
}

// Helper to return all lists as HTML partials
func returnAllLists(c *fiber.Ctx) error {
	lists, err := db.GetAllLists()
	if err != nil {
		return sendError(c, 500, "error.fetch_failed")
	}

	activeList, _ := db.GetActiveList()

	return c.Render("partials/lists_container", fiber.Map{
		"Lists":      lists,
		"ActiveList": activeList,
	}, "")
}

// ToggleShowCompleted toggles the show_completed setting for a list
func ToggleShowCompleted(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}

	list, err := db.ToggleListShowCompleted(id)
	if err != nil {
		return c.Status(500).SendString("Failed to toggle show completed")
	}

	// Broadcast to WebSocket clients
	BroadcastUpdate("list_updated", list)

	// Return the updated sections list
	sections, err := db.GetSectionsByList(id)
	if err != nil {
		return c.Status(500).SendString("Failed to fetch sections")
	}

	return c.Render("partials/sections_list", fiber.Map{
		"Sections":      sections,
		"ShowCompleted": list.ShowCompleted,
	}, "")
}

// sectionRenderMap builds the template data map for rendering a single section partial
func sectionRenderMap(section *db.Section) fiber.Map {
	return fiber.Map{
		"Section":       section,
		"Sections":      getSectionsForDropdown(),
		"ShowCompleted": db.GetShowCompletedForSection(section.ID),
	}
}

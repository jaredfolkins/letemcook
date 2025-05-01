package handlers

import (
	"fmt"
	"net/http"

	"github.com/jaredfolkins/letemcook/models"
)

func PatchCookbookToggle(c LemcContext) error {
	cb := &models.Cookbook{AccountID: c.UserContext().ActingAs.Account.ID}
	uuid := c.Param("uuid")
	toggle_type := c.Param("toggle_type")
	err := cb.ByUUID(uuid)
	if err != nil {
		c.AddErrorFlash("cookbook", "Cookbook not found")
		return c.NoContent(http.StatusFound)
	}

	var b bool
	switch toggle_type {
	case "deleted":
		if cb.IsDeleted {
			cb.IsDeleted = false
		} else {
			cb.IsDeleted = true
			b = true
		}

		err = cb.Update()
		if err != nil {
			c.AddErrorFlash(toggle_type, "deleted error")
			return c.NoContent(http.StatusNoContent)
		}

		c.AddSuccessFlash(toggle_type, fmt.Sprintf("deleted set to %t", b))
		return c.NoContent(http.StatusNoContent)
	case "published":
		if cb.IsPublished {
			cb.IsPublished = false
		} else {
			cb.IsPublished = true
			b = true
		}

		err = cb.Update()
		if err != nil {
			c.AddErrorFlash(toggle_type, "published error")
			return c.NoContent(http.StatusNoContent)
		}

		c.AddSuccessFlash(toggle_type, fmt.Sprintf("published set to %t", b))
		return c.NoContent(http.StatusNoContent)
	}

	return c.JSON(http.StatusBadRequest, "Invalid toggle type")
}

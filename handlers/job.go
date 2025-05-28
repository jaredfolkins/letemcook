package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/paths"
	"github.com/jaredfolkins/letemcook/util"
	"github.com/jaredfolkins/letemcook/views/partials"
	"github.com/jaredfolkins/letemcook/yeschef"
	"github.com/labstack/gommon/log"
	"github.com/reugn/go-quartz/quartz"
	"gopkg.in/yaml.v3"
)

const (
	ENV_PRIVATE_PREVIX = "LEMC_PRIVATE_"
	ENV_PUBLIC_PREFIX  = "LEMC_PUBLIC_"
)

func validateFormName(input string) error {
	pattern := `^[a-zA-Z0-9_]+$`
	matched, err := regexp.MatchString(pattern, input)
	if err != nil {
		return err
	}
	if !matched {
		return errors.New("input does not match the required pattern")
	}
	return nil
}

func GetAppJobStatus(c LemcContext) error {
	uuid := c.Param("uuid")
	pageid := c.Param("page")
	scope := c.Param("scope")
	userid := strconv.FormatInt(c.UserContext().ActingAs.Account.ID, 10)

	app, err := models.AppByUUIDAndAccountID(uuid, c.UserContext().ActingAs.Account.ID)
	if err != nil {
		c.AddErrorFlash("error", "app not found or permission denied")
		return c.NoContent(http.StatusNotFound)
	}

	jr := &yeschef.JobRecipe{
		JobType:  yeschef.JOB_TYPE_APP,
		UUID:     app.UUID,
		PageID:   pageid,
		Scope:    scope,
		UserID:   userid,
		Username: c.UserContext().ActingAs.Username,
		AppID:    fmt.Sprintf("%d", app.ID),
	}
	js := NewJobStatus(jr)
	jsView := partials.JobStatusView(uuid, pageid, scope, js.NowRunning, js.InRunning, js.EveryRunning)

	return HTML(c, jsView)
}

func GetCookbookJobStatus(c LemcContext) error {
	uuid := c.Param("uuid")
	pageid := c.Param("page")
	scope := strings.TrimSpace(c.Param("scope"))
	userid := strconv.FormatInt(c.UserContext().ActingAs.Account.ID, 10)

	cb := models.Cookbook{}
	err := cb.ByUUIDAndAccountID(uuid, c.UserContext().ActingAs.Account.ID)
	if err != nil {
		c.AddErrorFlash("error", "cookbook not found or permission denied: "+err.Error())
		return c.NoContent(http.StatusNotFound)
	}

	jr := &yeschef.JobRecipe{
		JobType:    yeschef.JOB_TYPE_COOKBOOK,
		UUID:       cb.UUID,
		PageID:     pageid,
		Scope:      scope,
		UserID:     userid,
		Username:   c.UserContext().ActingAs.Username,
		CookbookID: fmt.Sprintf("%d", cb.ID),
	}

	js := NewJobStatus(jr)
	jsView := partials.JobStatusView(uuid, pageid, scope, js.NowRunning, js.InRunning, js.EveryRunning)

	return HTML(c, jsView)
}

type JobStatus struct {
	NowRunning   bool
	InRunning    bool
	EveryRunning bool
}

func NewJobStatus(jr *yeschef.JobRecipe) *JobStatus {
	js := &JobStatus{}

	// Check if XoxoX is initialized
	if yeschef.XoxoX == nil || yeschef.XoxoX.RunningMan == nil {
		return js
	}

	// Check NOW jobs - both running and scheduled
	nowKey := yeschef.LemcJobKey(jr, yeschef.NOW_QUEUE)
	if yeschef.XoxoX.RunningMan.IsRunning(nowKey) {
		js.NowRunning = true
	} else if yeschef.XoxoX.NowQueue != nil {
		// Check if there's a scheduled NOW job
		jobKey := quartz.NewJobKey(nowKey)
		if job, err := yeschef.XoxoX.NowQueue.Get(jobKey); err == nil && job != nil {
			js.NowRunning = true
		}
	}

	// Check IN jobs - both running and scheduled
	inKey := yeschef.LemcJobKey(jr, yeschef.IN_QUEUE)
	if yeschef.XoxoX.RunningMan.IsRunning(inKey) {
		js.InRunning = true
	} else if yeschef.XoxoX.InQueue != nil {
		// Check if there's a scheduled IN job
		jobKey := quartz.NewJobKey(inKey)
		if job, err := yeschef.XoxoX.InQueue.Get(jobKey); err == nil && job != nil {
			js.InRunning = true
		}
	}

	// Check EVERY jobs - both running and scheduled
	everyKey := yeschef.LemcJobKey(jr, yeschef.EVERY_QUEUE)
	if yeschef.XoxoX.RunningMan.IsRunning(everyKey) {
		js.EveryRunning = true
	} else if yeschef.XoxoX.EveryQueue != nil {
		// Check if there's a scheduled EVERY job
		jobKey := quartz.NewJobKey(everyKey)
		if job, err := yeschef.XoxoX.EveryQueue.Get(jobKey); err == nil && job != nil {
			js.EveryRunning = true
		}
	}

	return js
}

func PutCookbookJob(c LemcContext) error {
	var env []string
	var username, pageid string
	var yaml_default models.YamlDefault
	var final_recipe models.Recipe
	var recipientUserIDs []int64

	username = c.UserContext().ActingAs.Username
	uuid := c.Param("uuid")
	page := c.Param("page")
	recipe := c.Param("recipe")
	view_type := c.Param("view_type")

	formValues, err := c.FormParams()
	if err != nil {
		c.AddErrorFlash("error", err.Error())
		return c.NoContent(http.StatusConflict)
	}

	cb := models.Cookbook{}
	err = cb.ByUUIDAndAccountID(uuid, c.UserContext().ActingAs.Account.ID)
	if err != nil {
		c.AddErrorFlash("error", err.Error())
		return c.NoContent(http.StatusConflict)
	}

	var scope string
	var isShared bool
	var http_file_download string
	originatingUserID := c.UserContext().ActingAs.ID
	pagei, err := strconv.Atoi(page)
	if err != nil {
		c.AddErrorFlash("error", "error parsing page")
		return c.NoContent(http.StatusConflict)
	}

	switch view_type {
	case SCOPE_YAML_TYPE_INDIVIDUAL:
		scope = SCOPE_YAML_TYPE_INDIVIDUAL
		err = yaml.Unmarshal([]byte(cb.YamlIndividual), &yaml_default)
		if err != nil {
			c.AddErrorFlash("error", "error parsing yaml")
			return c.NoContent(http.StatusConflict)
		}
		recipientUserIDs = []int64{originatingUserID}
		isShared = false
		http_file_download = fmt.Sprintf(paths.LockerDownloadPattern, cb.UUID, pagei, SCOPE_YAML_TYPE_INDIVIDUAL)
	case SCOPE_YAML_TYPE_SHARED:
		scope = SCOPE_YAML_TYPE_SHARED
		err = yaml.Unmarshal([]byte(cb.YamlShared), &yaml_default)
		if err != nil {
			c.AddErrorFlash("error", "error parsing yaml")
			return c.NoContent(http.StatusConflict)
		}
		isShared = true
		ids, err := models.GetUserIDsForSharedCookbook(cb.UUID)
		if err != nil {
			log.Errorf("Error getting user IDs for shared cookbook %s: %v", cb.UUID, err)
			c.AddErrorFlash("error", "Failed to determine recipients for shared job.")
			return c.NoContent(http.StatusInternalServerError)
		}
		recipientUserIDs = ids
		http_file_download = fmt.Sprintf(paths.LockerDownloadPattern, cb.UUID, pagei, SCOPE_YAML_TYPE_SHARED)
	default:
		c.AddErrorFlash("error", "view_type not found")
		return c.NoContent(http.StatusConflict)
	}

	yaml_default.UUID = cb.UUID

	for _, private := range yaml_default.Cookbook.Environment.Private {
		log.Debug(private)
		env = append(env, private)
	}

	for _, public := range yaml_default.Cookbook.Environment.Public {
		log.Debug(public)
		env = append(env, public)
	}

	for _, p := range yaml_default.Cookbook.Pages {
		if p.PageID == pagei {
			for _, r := range p.Recipes {
				if r.Name == recipe {
					for key, values := range formValues {
						// Process all form fields without requiring a prefix
						if validateFormName(key) != nil {
							c.AddErrorFlash("error", "error parsing form field names, invalid characters in: "+key)
							return c.NoContent(http.StatusConflict)
						}
						uppercasedFieldName := strings.ToUpper(key)
						for _, value := range values {
							addEnv := fmt.Sprintf("%s=%s", uppercasedFieldName, value)
							env = append(env, addEnv)
						}
					}

					env = append(env, "LEMC_STEP_ID=1")
					env = append(env, "LEMC_SCOPE="+scope)
					env = append(env, "LEMC_USER_ID="+fmt.Sprintf("%d", originatingUserID))
					env = append(env, "LEMC_USERNAME="+username)
					env = append(env, "LEMC_UUID="+cb.UUID)
					env = append(env, "LEMC_RECIPE_NAME="+util.AlphaNumHyphen(r.Name))
					env = append(env, "LEMC_PAGE_ID="+fmt.Sprintf("%d", p.PageID))
					env = append(env, fmt.Sprintf("LEMC_HTTP_DOWNLOAD_BASE_URL=%s", http_file_download))
					pageid = strconv.Itoa(p.PageID)
					final_recipe = r
					final_recipe.IsShared = isShared
				}
			}
		}
	}

	job := &yeschef.JobRecipe{
		JobType:          yeschef.JOB_TYPE_COOKBOOK,
		UUID:             cb.UUID,
		PageID:           pageid,
		CookbookID:       fmt.Sprintf("%d", cb.ID),
		UserID:           fmt.Sprintf("%d", originatingUserID),
		Username:         username,
		Env:              env,
		Scope:            scope,
		Recipe:           final_recipe,
		RecipientUserIDs: recipientUserIDs,
	}

	if missing, err := yeschef.CheckJobImages(job); err == nil && len(missing) > 0 {
		c.AddErrorFlash("error", "missing container images: "+strings.Join(missing, ", "))
		return c.NoContent(http.StatusConflict)
	} else if err != nil {
		c.AddErrorFlash("error", "failed to check images: "+err.Error())
		return c.NoContent(http.StatusConflict)
	}

	msg := "job submitted and monitor opened"

	err = yeschef.DoNow(job)
	if err != nil {
		if userErr := yeschef.GetUserVisibleError(err); userErr != nil {
			c.AddErrorFlash("error", userErr.Message)
		} else {
			c.AddErrorFlash("error", "Failed to start job: "+err.Error())
		}
		return c.NoContent(http.StatusConflict)
	}

	c.AddSuccessFlash("success", msg)
	return HTML(c, partials.OpenMonitorModal(cb.UUID, pageid, msg))
}

func PutAppJob(c LemcContext) error {
	var env []string
	var username, pageid string
	var yaml_default models.YamlDefault
	var final_recipe models.Recipe
	var recipientUserIDs []int64

	username = c.UserContext().ActingAs.Username
	uuid := c.Param("uuid")
	page := c.Param("page")
	recipe := c.Param("recipe")
	view_type := c.Param("view_type")

	formValues, err := c.FormParams()
	if err != nil {
		c.AddErrorFlash("error", err.Error())
		return c.NoContent(http.StatusConflict)
	}

	// Look up App first
	app, err := models.AppByUUIDAndAccountID(uuid, c.UserContext().ActingAs.Account.ID)
	if err != nil {
		log.Printf("Error fetching app by UUID %s for account %d: %v", uuid, c.UserContext().ActingAs.Account.ID, err)
		if errors.Is(err, sql.ErrNoRows) {
			c.AddErrorFlash("error", "app not found or permission denied")
			return c.NoContent(http.StatusNotFound)
		}
		c.AddErrorFlash("error", "Error retrieving app")
		return c.NoContent(http.StatusInternalServerError)
	}

	// Get associated Cookbook
	// hacky: we need to get the cookbook ID from the app
	CookbookPretendingToBeApp, err := models.CookbookByIDAndAccountID(app.CookbookID, c.UserContext().ActingAs.Account.ID)
	if err != nil {
		log.Printf("Error fetching cookbook by ID %d for account %d (from app %s): %v", app.CookbookID, c.UserContext().ActingAs.Account.ID, uuid, err)
		if errors.Is(err, sql.ErrNoRows) {
			c.AddErrorFlash("error", "Associated cookbook not found")
			return c.NoContent(http.StatusNotFound)
		}
		c.AddErrorFlash("error", "Error retrieving associated cookbook")
		return c.NoContent(http.StatusInternalServerError)
	}

	// Populate Cookbook struct with App details
	CookbookPretendingToBeApp.UUID = app.UUID
	CookbookPretendingToBeApp.Name = app.Name
	CookbookPretendingToBeApp.Description = app.Description
	CookbookPretendingToBeApp.YamlShared = app.YAMLShared
	CookbookPretendingToBeApp.YamlIndividual = app.YAMLIndividual

	var scope string
	var isShared bool
	var http_file_download string
	originatingUserID := c.UserContext().ActingAs.ID
	pagei, err := strconv.Atoi(page)
	if err != nil {
		c.AddErrorFlash("error", "error parsing page")
		return c.NoContent(http.StatusConflict)
	}

	switch view_type {
	case SCOPE_YAML_TYPE_INDIVIDUAL:
		scope = SCOPE_YAML_TYPE_INDIVIDUAL
		err = yaml.Unmarshal([]byte(CookbookPretendingToBeApp.YamlIndividual), &yaml_default)
		if err != nil {
			c.AddErrorFlash("error", "error parsing yaml")
			return c.NoContent(http.StatusConflict)
		}
		recipientUserIDs = []int64{originatingUserID}
		isShared = false

		http_file_download = fmt.Sprintf(paths.LockerDownloadPattern, CookbookPretendingToBeApp.UUID, pagei, SCOPE_YAML_TYPE_INDIVIDUAL)
	case SCOPE_YAML_TYPE_SHARED:
		scope = SCOPE_YAML_TYPE_SHARED
		err = yaml.Unmarshal([]byte(CookbookPretendingToBeApp.YamlShared), &yaml_default)
		if err != nil {
			c.AddErrorFlash("error", "error parsing yaml")
			return c.NoContent(http.StatusConflict)
		}
		isShared = true
		ids, err := models.GetUserIDsForSharedApp(CookbookPretendingToBeApp.UUID)
		if err != nil {
			log.Errorf("Error getting user IDs for shared app %s: %v", CookbookPretendingToBeApp.UUID, err)
			c.AddErrorFlash("error", "Failed to determine recipients for shared job.")
			return c.NoContent(http.StatusInternalServerError)
		}
		recipientUserIDs = ids
		http_file_download = fmt.Sprintf(paths.LockerDownloadPattern, CookbookPretendingToBeApp.UUID, pagei, SCOPE_YAML_TYPE_SHARED)
	default:
		c.AddErrorFlash("error", "view_type not found")
		return c.NoContent(http.StatusConflict)
	}

	yaml_default.UUID = CookbookPretendingToBeApp.UUID

	for _, private := range yaml_default.Cookbook.Environment.Private {
		log.Debug(private)
		env = append(env, private)
	}

	for _, public := range yaml_default.Cookbook.Environment.Public {
		log.Debug(public)
		env = append(env, public)
	}

	for _, p := range yaml_default.Cookbook.Pages {
		if p.PageID == pagei {
			for _, r := range p.Recipes {
				if r.Name == recipe {
					for key, values := range formValues {
						// Process all form fields without requiring a prefix
						if validateFormName(key) != nil {
							c.AddErrorFlash("error", "error parsing form field names, invalid characters in: "+key)
							return c.NoContent(http.StatusConflict)
						}
						uppercasedFieldName := strings.ToUpper(key)
						for _, value := range values {
							addEnv := fmt.Sprintf("%s=%s", uppercasedFieldName, value)
							env = append(env, addEnv)
						}
					}

					env = append(env, "LEMC_SCOPE="+scope)
					env = append(env, "LEMC_USER_ID="+fmt.Sprintf("%d", originatingUserID))
					env = append(env, "LEMC_USERNAME="+username)
					env = append(env, "LEMC_UUID="+CookbookPretendingToBeApp.UUID)
					env = append(env, "LEMC_RECIPE_NAME="+util.AlphaNumHyphen(r.Name))
					env = append(env, "LEMC_PAGE_ID="+fmt.Sprintf("%d", p.PageID))
					env = append(env, fmt.Sprintf("LEMC_HTTP_DOWNLOAD_BASE_URL=%s", http_file_download))
					pageid = strconv.Itoa(p.PageID)
					final_recipe = r
					final_recipe.IsShared = isShared
				}
			}
		}
	}

	job := &yeschef.JobRecipe{
		JobType:          yeschef.JOB_TYPE_APP,
		UUID:             CookbookPretendingToBeApp.UUID,
		PageID:           pageid,
		AppID:            fmt.Sprintf("%d", app.ID),
		UserID:           fmt.Sprintf("%d", originatingUserID),
		Username:         username,
		Env:              env,
		Scope:            scope,
		Recipe:           final_recipe,
		RecipientUserIDs: recipientUserIDs,
	}

	if missing, err := yeschef.CheckJobImages(job); err == nil && len(missing) > 0 {
		c.AddErrorFlash("error", "missing container images: "+strings.Join(missing, ", "))
		return c.NoContent(http.StatusConflict)
	} else if err != nil {
		c.AddErrorFlash("error", "failed to check images: "+err.Error())
		return c.NoContent(http.StatusConflict)
	}

	msg := "job submitted and monitor opened"

	err = yeschef.DoNow(job)
	if err != nil {
		if userErr := yeschef.GetUserVisibleError(err); userErr != nil {
			c.AddErrorFlash("error", userErr.Message)
		} else {
			c.AddErrorFlash("error", "Failed to start job: "+err.Error())
		}
		return c.NoContent(http.StatusConflict)
	}

	c.AddSuccessFlash("success", msg)
	return HTML(c, partials.OpenMonitorModal(CookbookPretendingToBeApp.UUID, pageid, msg))
}

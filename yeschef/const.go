package yeschef

import (
	"io/fs"
	"time"

	"github.com/jaredfolkins/letemcook/util"
)

const (
	STEP_ID           = "LEMC_STEP_ID=%d"
	LEMC_HTML_ID      = "LEMC_HTML_ID=uuid-%s-pageid-%s-scope-%s-html"
	LEMC_CSS_ID       = "LEMC_CSS_ID=uuid-%s-pageid-%s-scope-%s-style"
	LEMC_JS_ID        = "LEMC_JS_ID=uuid-%s-pageid-%s-scope-%s-script"
	NOW_QUEUE         = "now"
	IN_QUEUE          = "in"
	EVERY_QUEUE       = "every"
	PYTHON_UNBUFFERED = "PYTHONUNBUFFERED=1"
	LEMC_CSS_TRUNC    = "lemc.css.trunc;"
	LEMC_CSS_BUFFER   = "lemc.css.buffer;"
	LEMC_CSS_APPEND   = "lemc.css.append;"
	LEMC_HTML_TRUNC   = "lemc.html.trunc;"
	LEMC_HTML_BUFFER  = "lemc.html.buffer;"
	LEMC_HTML_APPEND  = "lemc.html.append;"
	LEMC_JS_EXEC      = "lemc.js.exec;"
	LEMC_JS_TRUNC     = "lemc.js.trunc;"
	LEMC_ERR          = "lemc.err;"
	LEMC_ENV          = "lemc.env;"
	OWNED_BY          = "LEMC"
	MAX_MESSAGE_SIZE  = 512
	JOB_TYPE_APP      = "app"
	JOB_TYPE_COOKBOOK = "cookbook"

	WRITE_WAIT  = 10 * time.Second
	PONG_WAIT   = 60 * time.Second
	PING_PERIOD = (PONG_WAIT * 9) / 10

	FILE_MODE fs.FileMode = util.DirPerm

	html_fn = "cache.html"
	css_fn  = "cache.css"
	js_fn   = "cache.js"
)

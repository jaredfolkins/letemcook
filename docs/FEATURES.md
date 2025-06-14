# Let'em Cook (LEMC) - Key Features

This document outlines the key features of LEMC, designed to help developers, LLMs, and agents understand how to utilize the system.

## Table of Contents
*   [Core Concept](#core-concept)
*   [Key Takeaways](#key-takeaways)
*   [Key Components](#key-components)
*   [Execution Environment](#execution-environment)
*   [Real-Time UI Communication (LEMC Verbs)](#real-time-ui-communication-lemc-verbs)
*   [Scheduling](#scheduling)
*   [Tech Stack Highlights](#tech-stack-highlights)
*   [Security](#security)
*   [Automatically Injected Environment Variables](#automatically-injected-environment-variables)
*   [Form Fields and User Input](#form-fields-and-user-input)
*   [Step Environment Variables](#step-environment-variables)
*   [Quick Start: Creating and Using a Local Script with Docker](#quick-start-creating-and-using-a-local-script-with-docker)
*   [Mounted File System in Containers](#mounted-file-system-in-containers)
*(This ToC can be expanded and refined)*

## Key Takeaways

*   **Automation Focus:** LEMC automates script execution ("recipes") in containerized environments.
*   **Developer-Centric:** Designed for developers to manage their own operational tasks ("Ops your Devs").
*   **UI Feedback:** Scripts communicate with a web UI in real-time using "LEMC verbs."
*   **Language Agnostic:** Run scripts in any language that can be containerized.
*   **Core Workflow:** Package scripts in Docker, define them as recipe steps in LEMC, and run via UI or schedule.

## Core Concept

LEMC (Let'em Cook) is an open-source tool for automating and executing predefined "recipes" (scripts) on demand, with results streamed live to a web interface, empowering developers to perform operational tasks. For a deeper understanding of the design principles and motivations behind LEMC, please see [PHILOSOPHY.md](PHILOSOPHY.md).

## Key Components

*   **Cookbooks**: Collections of related recipes (tasks).
*   **Recipes**: A defined sequence of one or more script-based steps to accomplish a specific task. Presented as runnable actions in the UI.
*   **Steps**: Individual scripts within a recipe. LEMC is language-agnostic; steps can be written in Bash, Python, Go, etc.

## Execution Environment

*   **Containerized**: Recipes run in Docker containers, ensuring consistent and isolated execution. Each recipe step typically uses a Docker image containing its code and dependencies.
*   **Language-Agnostic**: Scripts can be written in any language that can run in a container and print to standard output.

## Real-Time UI Communication (LEMC Verbs)

Scripts communicate with the LEMC UI by printing specially formatted strings (verbs) to standard output. The `yeschef` application (LEMC's backend) parses these commands to dynamically update the UI with HTML, CSS, or execute JavaScript, and to pass environment variables between steps.

For a detailed reference of all available LEMC Verbs and examples of their usage, including helper functions, please see [SCRIPT_UI_COMMUNICATION.md](SCRIPT_UI_COMMUNICATION.md).

**Brief Overview of Verb Categories:**
*   **`lemc.env;KEY=value`**: Sets environment variables for subsequent steps.
*   **`lemc.css.*`**: Verbs to manage CSS (append, truncate).
*   **`lemc.html.*`**: Verbs to manage HTML content (append, truncate).
*   **`lemc.js.*`**: Verbs to manage and execute JavaScript (truncate, execute).

**Note on `lemc.env`:** The LEMC backend collects `KEY=value` pairs from `lemc.env` outputs. These are then injected as environment variables into the container for the *next* step of the recipe.

## Scheduling

*   Recipes can be scheduled to run periodically (cron-like functionality) via the go-quartz library.
*   This allows for managed, recurring tasks with UI feedback and logging.

## Philosophy
*(This section has been integrated into the Core Concept and [PHILOSOPHY.md](PHILOSOPHY.md))*

## Tech Stack Highlights

*   **Backend**: Go (Golang) with the Echo web framework.
*   **Database**: SQLite.
*   **Frontend**: Server-side templating (Templ) with HTMX for interactivity.
*   **Styling**: Tailwind CSS.
*   **Containerization**: Docker.

## Security

*   Relies on Docker container isolation as the primary sandboxing mechanism.
*   Lightweight user/permission model suitable for small teams.
*   Admin account created on first launch; admin can manage users.
Access to the Docker socket is a requirement.

This feature set allows for flexible and powerful automation directly from scripts, with real-time updates to a web interface, making operational tasks more accessible and manageable for development teams.

## Automatically Injected Environment Variables

LEMC injects several environment variables into the step's container for context. Key variables include:

| Variable                      | Description                                                                                                | Example Value / Format                      |
| ----------------------------- | ---------------------------------------------------------------------------------------------------------- | ------------------------------------------- |
| `LEMC_STEP_ID`                | Current step number in the recipe.                                                                         | `1`, `2`                                    |
| `LEMC_SCOPE`                  | Job scope: "individual" (user-run) or "shared".                                                          | `individual`                                |
| `LEMC_USER_ID`                | Numerical ID of the initiating user.                                                                       | `101`                                       |
| `LEMC_USERNAME`               | Username of the initiating user.                                                                           | `jdoe`                                      |
| `LEMC_UUID`                   | UUID of the parent Cookbook/App.                                                                           | `ac72b1e9-0489-4b28-9df5-179c70419203`      |
| `LEMC_RECIPE_NAME`            | Name of the current recipe (sanitized).                                                                    | `my-awesome-recipe`                         |
| `LEMC_PAGE_ID`                | Numerical ID of the Cookbook page for the recipe.                                                          | `3`                                         |
| `LEMC_HTTP_DOWNLOAD_BASE_URL` | Base path for constructing download links. Files placed by a script into its `/lemc/public/` mounted directory are automatically made accessible via HTTP GET requests to a URL formed by this base path followed by the filename. The path structure typically includes context like UUID, page ID, and scope, and ends with `/filename/`. | `/lemc/locker/uuid/<UUID>/page/<PAGE_ID>/scope/<SCOPE>/filename/` |
| `LEMC_HTML_ID`                | Dynamically generated ID for the job's HTML output container in the UI.                                     | `uuid-JOB_UUID-pageid-PAGE_ID-scope-SCOPE-html` |
| `LEMC_CSS_ID`                 | Dynamically generated ID for the job's CSS `<style>` tag.                                                   | `uuid-JOB_UUID-pageid-PAGE_ID-scope-SCOPE-style` |
| `LEMC_JS_ID`                  | Dynamically generated ID for the job's JavaScript `<script>` area.                                          | `uuid-JOB_UUID-pageid-PAGE_ID-scope-SCOPE-script` |

**Form-Derived Variables**: For each form field defined in a recipe (e.g., a field named `My_Param` or `my-parameter` in the YAML), LEMC creates an environment variable. The HTML form input will be named using the field name directly (derived from your YAML definition, with spaces and hyphens converted to underscores, e.g., `My_Param` or `my_parameter`). In the container, the environment variable name will be the field name **converted to uppercase**.
    *   Example: If YAML field is `My_Param` (HTML form name `My_Param`), the resulting env var in the container is `MY_PARAM=value`.
    *   Example: If YAML field is `my_lower_param` (HTML form name `my_lower_param`), env var is `MY_LOWER_PARAM=value`.
    *   Example: If YAML field is `my-mixed-Param` (HTML form name `my_mixed_Param`), env var is `MY_MIXED_PARAM=value`.

**Note on `LEMC_HTTP_DOWNLOAD_BASE_URL`**: To form a complete URL, append the specific filename directly after this base path. The path resolves to a file within the job's public artifact store. For example, if `LEMC_HTTP_DOWNLOAD_BASE_URL` is `/lemc/locker/uuid/abc/page/1/scope/individual/filename/` and your file is `report.txt`, the full path used in a link would be `/lemc/locker/uuid/abc/page/1/scope/individual/filename/report.txt`.

These variables provide scripts with essential runtime information.

## Form Fields and User Input

LEMC supports dynamic form fields that allow users to provide input when executing recipes. These form inputs are automatically converted to environment variables accessible within the recipe steps.

### Form Field Types

LEMC supports the following form field types:
- `text`: Single-line text input
- `password`: Password input (masked)
- `textarea`: Multi-line text input
- `select`: Dropdown selection
- `radio`: Radio button selection

### Form Field Definitions

Form fields can be defined using two formats:

#### Legacy Format (Backward Compatible)
The original format uses a `name` field and `defaults` array where each item serves as both the display label and submitted value:

```yaml
form:
  - name: environment
    type: select
    defaults:
      - development
      - staging
      - production
  - name: username
    type: text
    defaults:
      - enter your username
```

#### New Format with Variable and Description
The enhanced format uses `variable` for the environment variable name, `description` for the display label, and `options` for separate labels and values:

```yaml
form:
  - variable: environment
    description: "Select deployment environment"
    type: select
    options:
      - label: "🔧 Development Environment"
        value: "dev"
      - label: "🚀 Staging Environment"
        value: "staging"
      - label: "🏭 Production Environment"
        value: "prod"
  - variable: log_level
    description: "Choose logging level"
    type: radio
    options:
      - label: "Debug (Verbose)"
        value: "debug"
      - label: "Info (Standard)"
        value: "info"
      - label: "Error (Critical Only)"
        value: "error"
```

#### Mixed Format
You can mix both formats within the same recipe:

```yaml
form:
  - variable: server_type
    description: "Select server type"
    type: select
    options:
      - label: "Database Server"
        value: "db"
      - label: "Web Server"
        value: "web"
  - name: region  # Legacy format
    type: radio
    defaults:
      - us-east-1
      - us-west-2
  - variable: admin_email
    description: "Administrator email"
    type: text
    defaults:
      - admin@example.com
```

### Environment Variable Conversion

- Form field variable names are converted to uppercase environment variables
- Spaces and hyphens in variable names are converted to underscores
- The **value** (not the label or description) is used as the environment variable value

**Examples:**
- Field `variable: deployment_env` with value `"prod"` → Environment variable `DEPLOYMENT_ENV=prod`
- Field `variable: log-level` with value `"debug"` → Environment variable `LOG_LEVEL=debug`

### Field Structure Reference

| Field | Required | Description |
|-------|----------|-------------|
| `variable` | Yes* | Environment variable name (preferred over `name`) |
| `name` | Yes* | Environment variable name (legacy, use `variable` instead) |
| `description` | No | Human-readable label displayed in the UI |
| `type` | Yes | Field type: `text`, `password`, `textarea`, `select`, `radio` |
| `options` | No | Array of label/value pairs for `select` and `radio` types |
| `defaults` | No | Default values or placeholder text |

*Either `variable` or `name` is required. `variable` takes precedence if both are present.

## Quick Start: Creating and Using a Local Script with Docker

For a step-by-step guide on creating your first recipe, including writing a script, containerizing it with Docker, and running it in LEMC, please refer to the [GETTING_STARTED.md](GETTING_STARTED.md) tutorial.

## Mounted File System in Containers

LEMC mounts several host directories into the container under `/lemc/` to allow scripts to access and store data. Standard mount points include:

| Mount Point      | Host Source (Conceptual)                                          | Purpose                                                                                                                                                                                          |
| ---------------- | ----------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `/lemc/public/`  | User-specific public directory (`cf.BindPerUserPublicDir`)          | For user-specific files, potentially web-accessible if served.                                                                                                                                   |
| `/lemc/private/` | User-specific private directory (`cf.BindPerUserPrivateDir`)        | For user-specific private files, not directly served.                                                                                                                                            |
| `/lemc/global/`  | UUID-specific 'locker' context directory (`cf.BindGlobalDir`)       | Global within `LEMC_UUID` scope (Cookbook/App 'locker'). For shared utilities/data relevant to that UUID's context.                                                                               |
| `/lemc/shared/`  | Common directory for shared recipes (`cf.BindSharedDir`)            | Mounted for "shared" recipes. For resources accessible to any job running that specific shared recipe.                                                                                           |

These mounts provide structured file system interaction for containerized scripts.

## Step Environment Variables

LEMC allows you to define custom environment variables for individual recipe steps. These variables are injected into the container environment when the step executes.

### Step Environment Field Definitions

Step environment variables can be defined using two formats:

#### Legacy Format (Backward Compatible)
The original format uses an `env` field:

```yaml
steps:
  - step: 1
    name: deploy application
    image: my-deployment-image:latest
    env:
      - DEPLOYMENT_ENV=production
      - LOG_LEVEL=info
    do: now
    timeout: 5.minutes
```

#### New Format (Recommended)
The enhanced format uses an `environment` field for better readability:

```yaml
steps:
  - step: 1
    name: deploy application
    image: my-deployment-image:latest
    environment:
      - DEPLOYMENT_ENV=production
      - LOG_LEVEL=info
      - SERVICE_NAME=my-application
    do: now
    timeout: 5.minutes
```

#### Mixed Format
You can mix both formats within the same recipe (different steps):

```yaml
steps:
  - step: 1
    name: legacy step
    image: legacy-image:latest
    env:
      - LEGACY_VAR=value
    do: now
    timeout: 5.minutes
  - step: 2
    name: modern step
    image: modern-image:latest
    environment:
      - MODERN_VAR=value
      - ENHANCED_CONFIG=enabled
    do: now
    timeout: 5.minutes
```

### Environment Variable Precedence

Environment variables are combined from multiple sources in the following order (later sources override earlier ones):

1. **Cookbook Environment**: Variables defined at the cookbook level (`public` and `private`)
2. **Form Input**: Variables derived from user form inputs
3. **System Variables**: Automatically injected LEMC variables (see [Automatically Injected Environment Variables](#automatically-injected-environment-variables))
4. **Step Environment**: Variables defined in the step's `environment` or `env` field

### Step Environment Field Reference

| Field | Format | Description |
|-------|--------|-------------|
| `env` | Legacy | Array of environment variables (use `environment` instead) |
| `environment` | Recommended | Array of environment variables with better readability |

**Note**: If both `env` and `environment` are present in the same step, `environment` takes precedence.
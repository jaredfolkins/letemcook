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

LEMC (Let'em Cook) is an open-source tool for automating and executing predefined "recipes" (scripts) on demand, with results streamed live to a web interface. It empowers developers to perform operational tasks ("Ops your Devs") by packaging scripts into containers and providing a UI for execution. For a deeper understanding of the design principles and motivations behind LEMC, please see [[PHILOSOPHY.md]].

## Key Components

*   **Cookbooks**: Collections of related recipes (tasks).
*   **Recipes**: A defined sequence of one or more script-based steps to accomplish a specific task. Presented as runnable actions in the UI.
*   **Steps**: Individual scripts within a recipe. LEMC is language-agnostic; steps can be written in Bash, Python, Go, etc.

## Execution Environment

*   **Containerized**: Recipes run in Docker containers, ensuring consistent and isolated execution. Each recipe step typically uses a Docker image containing its code and dependencies.
*   **Language-Agnostic**: Scripts can be written in any language that can run in a container and print to standard output.

## Real-Time UI Communication (LEMC Verbs)

Scripts communicate with the LEMC UI by printing specially formatted strings (verbs) to standard output. The `yeschef` application (LEMC's backend) parses these commands.

**LEMC Verb Reference:**

| Verb                    | Description                                                                 | Example                                                             |
| ----------------------- | --------------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `lemc.env;KEY=value`    | Sets `KEY=value` as an env var for subsequent steps in the same recipe.     | `echo "lemc.env;MY_VAR=hello"`                                        |
| `lemc.css.buffer;CSS`   | Appends CSS to a temporary buffer. (Effectively `lemc.css.append;`)         | `echo "lemc.css.buffer;body { color: blue; }"`                      |
| `lemc.css.trunc;CSS`    | Clears existing CSS and replaces it with `CSS`.                             | `echo "lemc.css.trunc;.new { font-weight: bold; }"`                 |
| `lemc.css.append;CSS`   | Appends `CSS` to the current CSS and renders.                               | `echo "lemc.css.append;p { margin: 5px; }"`                         |
| `lemc.html.buffer;HTML` | Appends HTML to a temporary buffer. (Effectively `lemc.html.append;`)       | `echo "lemc.html.buffer;<h2>Section</h2>"`                         |
| `lemc.html.trunc;HTML`  | Clears existing HTML and replaces it with `HTML`.                           | `echo "lemc.html.trunc;<div>Main Content</div>"`                    |
| `lemc.html.append;HTML` | Appends `HTML` to the current HTML and renders.                             | `echo "lemc.html.append;<p>More details...</p>"`                     |
| `lemc.js.trunc;JS`      | Clears existing JavaScript, replaces it with `JS`, then executes `JS`.      | `echo "lemc.js.trunc;console.log('JS Reset');"`                    |
| `lemc.js.exec;JS`       | Effectively same as `lemc.js.trunc;` - replaces and executes JavaScript.    | `echo "lemc.js.exec;alert('Hello from LEMC!');"`                    |

**Note on `lemc.env`:** The LEMC backend collects `KEY=value` pairs from `lemc.env` outputs. These are then injected as environment variables into the container for the *next* step of the recipe.

**Usage:**
Use `echo` (shell) or `print` (Python, etc.) to send these commands. Helper functions are recommended for cleaner scripts (see `AGENT.MD` for examples).

## Scheduling

*   Recipes can be scheduled to run periodically (cron-like functionality) via the go-quartz library.
*   This allows for managed, recurring tasks with UI feedback and logging.

## Philosophy
*(This section has been integrated into the Core Concept and PHILOSOPHY.md)*

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

**Form-Derived Variables**: For each form field defined in a recipe (e.g., a field named `My_Param` or `my-parameter` in the YAML), LEMC creates an environment variable. The HTML form input will be named `LEMC_FIELD_FORM_NAME_HERE` (where `FORM_NAME_HERE` is derived from your YAML definition, with spaces and hyphens converted to underscores, e.g., `LEMC_FIELD_My_Param` or `LEMC_FIELD_my_parameter`). In the container, the environment variable name will **retain the `LEMC_FIELD_` prefix**, and the part of the name following the prefix (derived from the YAML definition) will be **converted to uppercase**.
    *   Example: If YAML field is `My_Param` (HTML form name `LEMC_FIELD_My_Param`), the resulting env var in the container is `LEMC_FIELD_MY_PARAM=value`.
    *   Example: If YAML field is `my_lower_param` (HTML form name `LEMC_FIELD_my_lower_param`), env var is `LEMC_FIELD_MY_LOWER_PARAM=value`.
    *   Example: If YAML field is `my-mixed-Param` (HTML form name `LEMC_FIELD_my_mixed_Param`), env var is `LEMC_FIELD_MY_MIXED_PARAM=value`.

**Note on `LEMC_HTTP_DOWNLOAD_BASE_URL`**: To form a complete URL, append the specific filename directly after this base path. The path resolves to a file within the job's public artifact store. For example, if `LEMC_HTTP_DOWNLOAD_BASE_URL` is `/lemc/locker/uuid/abc/page/1/scope/individual/filename/` and your file is `report.txt`, the full path used in a link would be `/lemc/locker/uuid/abc/page/1/scope/individual/filename/report.txt`.

These variables provide scripts with essential runtime information.

## Quick Start: Creating and Using a Local Script with Docker

This section provides a brief overview of packaging a local script into a Docker container for use in a LEMC recipe. (Consider moving this to a dedicated `QUICK_START.md` or `TUTORIAL.md` for more detail).

**Core Steps:**
1.  **Create Script:** Write your task script (e.g., `my_task.sh`), using LEMC verbs for UI output.
2.  **Dockerfile:** Create a `Dockerfile` to package your script and its dependencies (e.g., `FROM alpine`, `COPY my_task.sh`, `CMD ["/app/my_task.sh"]`).
3.  **Build Image:** Build your Docker image (`docker build -t my-lemc-script:latest .`). Optionally, tag and push to a registry.
4.  **LEMC Recipe:** In the LEMC UI, create/update a recipe step to use your Docker image name.
5.  **Run:** Execute the recipe from a LEMC App.

LEMC runs your script in the container, streaming output to the UI.

## Mounted File System in Containers

LEMC mounts several host directories into the container under `/lemc/` to allow scripts to access and store data. Standard mount points include:

| Mount Point      | Host Source (Conceptual)                                          | Purpose                                                                                                                                                                                          |
| ---------------- | ----------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `/lemc/public/`  | User-specific public directory (`cf.BindPerUserPublicDir`)          | For user-specific files, potentially web-accessible if served.                                                                                                                                   |
| `/lemc/private/` | User-specific private directory (`cf.BindPerUserPrivateDir`)        | For user-specific private files, not directly served.                                                                                                                                            |
| `/lemc/global/`  | UUID-specific 'locker' context directory (`cf.BindGlobalDir`)       | Global within `LEMC_UUID` scope (Cookbook/App 'locker'). For shared utilities/data relevant to that UUID's context.                                                                               |
| `/lemc/shared/`  | Common directory for shared recipes (`cf.BindSharedDir`)            | Mounted for "shared" recipes. For resources accessible to any job running that specific shared recipe.                                                                                           |

These mounts provide structured file system interaction for containerized scripts.

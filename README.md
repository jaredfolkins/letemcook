<p align="center">
  <img src="logo.png" alt="LEMC Logo" width="400"/>
</p>

# Let'em Cook! 🔥 (LEMC)

## What problem is `LEMC` solving?

Have you ever found yourself thinking 
- _"I wish I could clone myself I'm oversubscribed!"_
- _"We have such a small team of coders & hackers, how could I force multiply them?"_
- _"I wonder if I could even leverage my boss or my customer to act as an extra set of hands?"_ 
- _"If only a button existed that could be clicked that would just do-the-thing..."_ 

Then `LEMC` is made for you!

## The `LEMC` thesis

In my experience when working on a team of developers or hackers, they often need to perform deterministic tasks, on non-deterministic schedules, and then communicate the results in a streamlined fashion.

Far too often these tasks fall under the domain of DevOps and so something like `Rundeck`, `Jenkins`, `GitHub Actions`, `GitLab CI/CD`, `Ansible AWX/Tower`, `Puppet Bolt`, `SaltStack`, `Chef Automate`, `Argo Workflows`, or `Apache Airflow` is implemented to help manage all-the-things. Unfortunately, these pieces of software can take a lot of support and tend to require a DevOps team with specialized knowledge as they are built for the enterprise market. This results in DevOps not acting as an extension of the team but rather its very own department. This causes a lot of friction, creating organizational drag, thus limiting the speed at which the team can ship.  

`LEMC` is built with the belief that in the age of vibe-coding, developers will out-pace their competition when they own their operations. Ultimately `LEMC` aims to help your organization "ops their devs."

## How does `LEMC` help?

`LEMC` works to free siloed code or business logic. It's the type of code that gets core business work done but tends to sit on someone's computer running under their desk or may need that "special engineer" around to run it manually. And when said engineer is out-of-office, suddenly the organization is screwed.

`LEMC` allows anyone on your team to take their siloed code or lone-wolf scripts, wrap them in a container, and quickly empower their team to get visual results streamed to the browser right from inside the container at the click of a button. It does this with a few special verbs and the most used programming functions of all time, `print` or `echo`. 

`LEMC` is built in anticipation of AI-assisted code generation which helps fast moving teams build and innovate quickly. Forsaking many modern and GUI-heavy solutions, `LEMC` is a **language first solution**. This is perfect for LLM vibe-coding sessions. 

#### At a high level how does `LEMC` accomplish this?

- **Cookbooks & Recipes:** `LEMC` is a web application with a lightweight security model for managing "cookbooks," which are collections of "recipes." Recipes can be executed on demand via a button click or scheduled to run periodically.
- **Script-Based Steps:** Each step within a recipe is a straightforward script, compatible with various languages (Bash, Python, Go, Ruby, Perl) or DSLs (Terraform, Ansible).
- **Containerized Workflow:**
    - Recipe scripts are packaged and distributed as container images using a Docker registry.
    - When a recipe is triggered, `LEMC` ensures the correct container image version is available locally (pulling it if necessary).
    - It then executes the container, automatically passing environment variables to manage state across steps or executions.
- **Real-time Feedback:** Recipe execution results are streamed live to the user's web browser via WebSockets and displayed within the `LEMC` UI.
- **Simplified Development & AI:** This container-centric approach simplifies recipe development and troubleshooting (using test harnesses) and provides a well-defined, small context when leveraging AI assistance more effectively.

Below I'll offer a tutorial that will use a real-world use case.

## Technology Stack

*   **Backend:** Go (Golang 1.23.0)
*   **Web Framework:** [Echo](https://echo.labstack.com/)
*   **Templating:** [Templ](https://templ.guide/)
*   **Frontend Interaction:** [HTMX](https://htmx.org/)
*   **Database:** SQLite
*   **Database Interaction:** [sqlx](https://github.com/jmoiron/sqlx), [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
*   **Database Migrations:** [pressly/goose](https://github.com/pressly/goose)
*   **Scheduling:** [go-quartz](https://github.com/reugn/go-quartz)
*   **Container Interaction:** Docker SDK
*   **Realtime:** WebSockets ([gorilla/websocket](https://github.com/gorilla/websocket))
*   **Containerization:** Docker, Docker Compose

## LEMC Core Development Requirements

To contribute to or develop `LEMC`, you'll need:

*   **Go:** Version 1.23.0 (check `.go-version` or `.goenv`).
*   **Docker & Docker Compose:** For running the application containerized and executing recipes.
*   **Access to a Docker Daemon Socket:** Required for `LEMC`'s core container interaction features (default `unix:///var/run/docker.sock`).
*   **[air](https://github.com/cosmtrek/air):** Recommended for live reloading during development (`go install github.com/cosmtrek/air@latest`).
*   **[templ](https://templ.guide/):** Required for compiling `.templ` files into Go code (`go install github.com/a-h/templ/cmd/templ@latest`).
*   **Git:** For version control.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

*   Go (version specified in `.go-version`)
*   Docker & Docker Compose (Optional, for running via container)
*   Access to a Docker daemon socket (required for recipe execution, default: `unix:///var/run/docker.sock`)

### Installation & Running

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/jaredfolkins/letemcook.git
    cd letemcook
    ```

2.  **Build the binary (Required for Manual / Live Reload):**
    ```bash
    go build -o ./tmp/lemc ./main.go
    ```

#### Manual Execution

*   Run the compiled binary directly:
    ```bash
    ./tmp/lemc
    ```
*   On the first run, LEMC will initialize the `data/` directory, create `data/.env`, run migrations, and start the server (default: `http://localhost:5362`).

#### Development (Live Reload)

*   Ensure you have `air` installed (`go install github.com/cosmtrek/air@latest`).
*   Make sure the binary is built (step 2 above).
*   Run `air` in the project root:
    ```bash
    air
    ```
*   `air` will monitor your Go files and automatically rebuild and restart the application on changes.

#### Docker Compose

*   Build and run the application using Docker Compose:
    ```bash
    docker-compose up --build
    ```
*   This uses the `Dockerfile` and `docker-compose.yml`.
*   The `data` directory and the Docker socket are automatically mounted into the container.

#### Initial Setup (After starting LEMC using any method)

*   Open your web browser and navigate to `http://localhost:5362`.
*   You will be redirected to the `/setup` page to create the initial administrator account.

## Quick Start: Your First Recipe 🔥

This guide will walk you through creating a simple "Hello World" recipe that also displays the current date from within its container.

### 1. Create Your Script (`my_recipe.sh`)

Create a file named `my_recipe.sh` in a new directory (e.g., `my-first-recipe/`) with the following content:

```bash
#!/bin/sh

echo "lemc.html.buffer; <h1>Hello from my LEMC recipe!</h1>"
echo "lemc.html.buffer; <p>The current date and time in the container is: <strong>$(date)</strong></p>"
echo "lemc.html.append;"
```

This script uses special `lemc.` prefixed commands:
*   `lemc.html.buffer; <HTML_CONTENT>`: Streams the `<HTML_CONTENT>` to the LEMC UI.
*   `lemc.html.append;`: Signals that a block of HTML has been completely sent and can be appended to the display.

Make the script executable:
```bash
chmod +x my_recipe.sh
```

### 2. Create a Dockerfile

In the same `my-first-recipe/` directory, create a file named `Dockerfile` with the following content:

```dockerfile
FROM alpine:latest

# Copy the script into the image
COPY my_recipe.sh /app/my_recipe.sh

# Set the working directory
WORKDIR /app

# Make the script executable (though already done, good practice in Dockerfile)
RUN chmod +x my_recipe.sh

# Command to run when the container starts
CMD ["/app/my_recipe.sh"]
```

This `Dockerfile` does the following:
*   Uses the lightweight `alpine:latest` base image.
*   Copies your `my_recipe.sh` script into the `/app/` directory in the image.
*   Sets `/app/` as the working directory.
*   Ensures the script is executable.
*   Specifies `my_recipe.sh` as the command to run when the container starts.

### 3. Build Your Docker Image

Navigate into your `my-first-recipe/` directory in the terminal and run the following command to build your Docker image:

```bash
docker build -t my-first-recipe:latest .
```
This command builds an image and tags it as `my-first-recipe:latest`.

**(Optional but Recommended for Sharing/Production)**
If you plan to use a Docker registry (like Docker Hub, GitLab Container Registry, etc.), you should tag and push your image:
```bash
# Tag the image with your registry username/namespace
docker tag my-first-recipe:latest your-registry-username/my-first-recipe:latest

# Push the image to the registry
docker push your-registry-username/my-first-recipe:latest
```
Replace `your-registry-username` with your actual username or namespace.

### 4. Add the Recipe in the LEMC UI

1.  **Open LEMC:** Navigate to your LEMC instance in your web browser (e.g., `http://localhost:5362`).
2.  **Go to Cookbooks:** Find the "Cookbooks" section in the navigation.
3.  **Create/Select a Cookbook:**
    *   If you don't have a cookbook, create a new one (e.g., "My Test Cookbook").
    *   Otherwise, select an existing cookbook where you want to add your recipe.
4.  **Add New Recipe:**
    *   Within the cookbook, find the option to add a new recipe.
    *   **Recipe Name:** Give your recipe a name (e.g., "My Hello World").
    *   **Description:** (Optional) Add a short description.
    *   **Image Name (Crucial):**
        *   If you pushed to a registry: `your-registry-username/my-first-recipe:latest`.
        *   If you built the image locally and did not push: `my-first-recipe:latest`. LEMC will attempt to use the local image if it can find it via the Docker daemon.
    *   **Timeout:** Set a timeout (e.g., `1.minute`).
    *   **`do` field**: For this simple example, ensure the `do` field for the step is set to `now`.
    *   Leave other fields at their default for now.
5.  **Save the Recipe.**

### 5. Run Your Recipe!

*   Once the recipe is saved, you should see it listed.
*   Click the "Run" (or similar) button next to your new recipe.
*   LEMC will pull the image (if not available locally and a full registry path was provided) and then run the container.
*   You should see the output ("Hello from my LEMC recipe!" and the current date/time) streamed to the UI.

Congratulations! You've created and run your first LEMC recipe. From here, you can explore more complex scripts, multi-step recipes, and other LEMC features.
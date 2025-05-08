<p align="center">
  <img src="media/lemc-readme-logo.png" alt="Let'em Cook! Logo" width="400"/>
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


## Special Thanks

I just wanted to say thanks to [Ed Skoudis](https://x.com/edskoudis) and the [CounterHack.com](https://www.counterhack.com/) team for always encouraging me to push myself! To allow for personal time and space to educate and innovate.

Ed you are truly a wonderful man and I'm thankful you are in my life.

<p align="left">
  <img src="media/counter-hack-white.png" alt="CounterHack Logo" width="125"/>
</p>


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
docker build -t lemc-my-first-recipe:latest .
```
This command builds an image and tags it as `lemc-my-first-recipe:latest`. It's a good practice to prefix your LEMC-related images with `lemc-` for easier identification.

**(Optional but Recommended for Sharing/Production)**
If you plan to use a Docker registry (like Docker Hub, GitLab Container Registry, etc.), you should tag and push your image:
```bash
# Tag the image with your registry, e.g., example.com/your-repo/
docker tag lemc-my-first-recipe:latest example.com/your-username/lemc-my-first-recipe:latest

# Push the image to the registry
docker push example.com/your-username/lemc-my-first-recipe:latest
```
Replace `example.com/your-username/` with your actual registry path and username/namespace.

### 4. Define Your Recipe in a LEMC Cookbook

1.  **Open LEMC:** Navigate to your LEMC instance in your web browser (e.g., `http://localhost:5362`).
2.  **Go to Cookbooks:** Find the "Cookbooks" section in the navigation.
3.  **Create or Select a Cookbook:**
    *   If you don't have a cookbook, create a new one (e.g., "My Test Cookbook"). Follow the UI prompts, which may involve clicking a "+" symbol or similar action to create a new cookbook.
    *   Otherwise, select an existing cookbook to which you want to add your new recipe.
4.  **Add Your New Recipe to the Cookbook:**
    *   Within the selected cookbook's interface, find the option to "Add New Recipe" or similar.
    *   **Recipe-Level Information:**
        *   **Recipe Name:** Give your recipe a unique name within the cookbook (e.g., "My Hello World").
        *   **Description:** (Optional) Add a short description of what your recipe does.
    *   **Details for the First Step (as recipes consist of one or more steps):**
        *   **Image Name:** This is the Docker image the step will run (e.g., `lemc-my-first-recipe:latest` if built locally, or `example.com/your-username/lemc-my-first-recipe:latest` if pushed to a registry). LEMC will use the Docker daemon to find/pull this image. Consider prefixing your LEMC image names with `lemc-` for better organization.
        *   **Timeout:** Specify how long this step is allowed to run (e.g., `1.minute`).
        *   **`do` field:** For this initial step, set this to `now`. This ensures the step executes when the recipe is manually triggered from an App.
    *   Leave other fields (such as those for form inputs or additional steps) at their default values for this simple example.
5.  **Save the Recipe.** This action saves the recipe definition within the cookbook. Ensure any overall cookbook changes are also saved if required by the UI.

### 5. Create a LEMC App from Your Cookbook

Now that your recipe is defined in a cookbook, you need to create an "App." An App is an instance of a cookbook, allowing you to run its recipes.

1.  **Go to Apps:** In the LEMC navigation bar, click on "Apps."
2.  **Initiate New App Creation:** On the "Apps" page, click the "+" button (as seen in the UI screenshot) to start creating a new App.
3.  **Configure the App:**
    *   **App Name:** Give your App a descriptive name (e.g., "My First Test App").
    *   **Select Cookbook:** From the available options, choose the cookbook that contains the "My Hello World" recipe you just defined.
    *   Fill in any other required App-specific details as prompted.
4.  **Save the App.** After saving, you should see your new App listed on the "Apps" page.

### 6. Run Your Recipe via the App!

*   **Open Your App:** From the "Apps" page, find your newly created App (e.g., "My First Test App"). Click on its name or the associated "App" button to open its interface.
*   **Navigate to the Correct Page/Tab:** Within the App's interface, you will see different pages or tabs (e.g., "Hello World Page," "Kitchen Sink," as shown in your screenshot). Click on the page/tab that contains the recipe you want to run. For this Quick Start, this would be the page where you expect your "My Hello World" recipe to be.
*   **Locate and Execute Your Recipe:** On the correct page, find your "My Hello World" recipe. It will likely be represented by a button labeled with the recipe name (e.g., a "hello world" button as shown in the screenshot under a recipe description like "basic hello world lemc example"). Click this button to execute the recipe.
*   LEMC will now pull the Docker image (if it's not already available locally and a full registry path was provided) and then execute the container.
*   You should see the output ("Hello from my LEMC recipe!" and the current date/time) streamed directly to the UI within the App's context.

Congratulations! You've defined a recipe within a cookbook, created an App to instance that cookbook, and successfully run your first LEMC recipe.

## Sponsors

???

---

Let'em Cook!, LEMC, and all associated code and assets are the property of Jared Folkins. Copyright © 2024 Jared Folkins. All rights reserved.
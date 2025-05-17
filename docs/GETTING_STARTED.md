# Getting Started

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
*   On the first run, LEMC will initialize the `data/<env>/` directory (where `<env>` is your `LEMC_ENV`), create `data/<env>/.env`, run migrations, and start the server using the port configured for your environment (default development port: `http://localhost:5362`).

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

*   Open your web browser and navigate to `http://localhost:<PORT>` (development defaults to `5362`).
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

1.  **Open LEMC:** Navigate to your LEMC instance in your web browser (e.g., `http://localhost:5362` if using the default development port).
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

# How does `LEMC` accomplish this?

- **Cookbooks & Recipes:** `LEMC` is a web application with a lightweight security model for managing "cookbooks," which are collections of "recipes." Recipes can be executed on demand via a button click or scheduled to run periodically.
- **Script-Based Steps:** Each step within a recipe is a straightforward script, compatible with various languages (Bash, Python, Go, Ruby, Perl) or DSLs (Terraform, Ansible).
- **Containerized Workflow:**
    - Recipe scripts are packaged and distributed as container images using a Docker registry.
    - When a recipe is triggered, `LEMC` ensures the correct container image version is available locally (pulling it if necessary).
    - It then executes the container, automatically passing environment variables to manage state across steps or executions.
- **Real-time Feedback:** Recipe execution results, including HTML content generated via simple `lemc` prefixed verbs (e.g., `echo "lemc.html.buffer; <h1>Update!</h1>"`), which are then streamed live to the user's web browser via WebSockets.
- **Simplified Development + Ai:** The container-centric approach makes it remarkably easy for developers to build and troubleshoot recipes locally. Once satisfied, they can push their container to a shared Docker registry, allowing for straightforward validation and updates by the team. This well-defined, language first, smaller context is also ideal for leveraging AI assistance more effectively.


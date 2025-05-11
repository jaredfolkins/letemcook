### How does `LEMC` accomplish this?

- **Cookbooks & Recipes:** `LEMC` is a web application with a lightweight security model for managing "cookbooks," which are collections of "recipes." Recipes can be executed on demand via a button click or scheduled to run periodically.
- **Script-Based Steps:** Each step within a recipe is a straightforward script, compatible with various languages (Bash, Python, Go, Ruby, Perl) or DSLs (Terraform, Ansible).
- **Containerized Workflow:**
    - Recipe scripts are packaged and distributed as container images using a Docker registry.
    - When a recipe is triggered, `LEMC` ensures the correct container image version is available locally (pulling it if necessary).
    - It then executes the container, automatically passing environment variables to manage state across steps or executions.
- **Real-time Feedback:** Recipe execution results, including HTML content generated via simple `lemc` prefixed verbs (e.g., `echo "lemc.html.buffer; <h1>Update!</h1>"`), which are then streamed live to the user's web browser via WebSockets.
- **Simplified Development + Ai:** The container-centric approach makes it remarkably easy for developers to build and troubleshoot recipes locally. Once satisfied, they can push their container to a shared Docker registry, allowing for straightforward validation and updates by the team. This well-defined, language first, smaller context is also ideal for leveraging AI assistance more effectively.

### Is LEMC a framework?

Not exactly. Let’em Cook is closer to a lightweight workflow-automation **platform** than a traditional software framework.

* **Frameworks** (like Django or React) embed themselves in your code. You write components that run *inside* their life-cycle.
* **Let’em Cook** sits **outside** your application code. You package each step of a workflow as a container (or script), then describe how those containers chain together in a YAML recipe. At run time the LEMC engine pulls the images, wires the steps, streams their output, and enforces RBAC.

Think of it as:

* **Orchestrator**: Spins up the right containers, handles retries, timeouts, fallbacks.
* **UI generator**: Reads the “verbs” you print (now, in, every, etc.) and turns them into buttons, logs, and dashboards without extra frontend work.
* **Distribution format**: Recipes plus docs and assets can be exported, shared, or version-controlled like any artifact.

So while it does give you conventions and helper libraries, you don’t *build* your software inside LEMC. Instead you plug your existing containers or scripts into it and let the platform do the orchestration.


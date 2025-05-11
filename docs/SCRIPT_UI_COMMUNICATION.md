# LEMC UI Communication Protocol

This section describes how containers (scripts running inside Docker containers) can communicate with the `yeschef` application to update environment variables, CSS, HTML, and JavaScript content for the user interface.

Communication is achieved by printing specially formatted strings to standard output. The `yeschef` application listens to the container's output stream and parses these commands.

## LEMC Verbs

The following "verbs" (prefixes) can be used with `echo` or `print` statements within your container scripts:

1.  **`lemc.env;`**: Sets an environment variable for the current step with the intent to pass the environment variable to forward for use with the next step.
    *   **Action**: Appends the provided string to the job's environment variables.
    *   **Example**: `echo "lemc.env;MY_VARIABLE=hello world"`

2.  **`lemc.css.buffer;`**: Buffers CSS content.
    *   **Action**: Appends the provided CSS string to the current CSS content for the view. *Behaves like `lemc.css.append;`.*
    *   **Example**: `echo "lemc.css.buffer;body { font-family: Arial, sans-serif; }"`

3.  **`lemc.css.trunc;`**: Truncates and replaces CSS content.
    *   **Action**: Clears any existing CSS for the view and replaces it with the provided CSS string.
    *   **Example**: `echo "lemc.css.trunc;.title { color: blue; }"`

4.  **`lemc.css.append;`**: Appends CSS content.
    *   **Action**: Appends the provided CSS string to the current CSS content for the view.
    *   **Example**: `echo "lemc.css.append;p { margin-bottom: 10px; }"`

5.  **`lemc.html.buffer;`**: Buffers HTML content.
    *   **Action**: Appends the provided HTML string to the current HTML content for the view. *Behaves like `lemc.html.append;`.*
    *   **Example**: `echo "lemc.html.buffer;<h1>Welcome</h1>"`

6.  **`lemc.html.trunc;`**: Truncates and replaces HTML content.
    *   **Action**: Clears any existing HTML for the view and replaces it with the provided HTML string.
    *   **Example**: `echo "lemc.html.trunc;<div>New section</div>"`

7.  **`lemc.html.append;`**: Appends HTML content.
    *   **Action**: Appends the provided HTML string to the current HTML content for the view.
    *   **Example**: `echo "lemc.html.append;<p>Additional details.</p>"`

8.  **`lemc.js.trunc;`**: Truncates and replaces JavaScript content.
    *   **Action**: Clears any existing JavaScript for the view and replaces it with the provided JavaScript string. This new JavaScript will then be executed on the client-side.
    *   **Example**: `echo "lemc.js.trunc;console.log('JavaScript has been reset and executed.');"`

9.  **`lemc.js.exec;`**: Executes JavaScript by replacing content.
    *   **Action**: Clears any existing JavaScript for the view and replaces it with the provided JavaScript string. This new JavaScript will then be executed on the client-side. *Effectively the same as `lemc.js.trunc;` based on current `yeschef/container.go` implementation.*
    *   **Example**: `echo "lemc.js.exec;alert('This JavaScript was executed!');"`

## Using Verbs

To use these verbs, simply `echo` or `print` (depending on your script's language) the command string. The `yeschef` application will intercept this output.

**Example (shell script):**
'''shell
#!/bin/sh
echo "lemc.html.trunc;<h1>My Dynamic Page</h1>"
echo "lemc.css.append;body { background-color: #eee; }"
echo "lemc.js.exec;console.log('Page updated by container.');"
echo "lemc.env;JOB_STATUS=completed"
'''

## Suggestion: Helper Functions

To make your agent scripts cleaner and less prone to typos, consider creating helper functions within your scripts or a shared library if your agent environment supports it.

These helper functions would encapsulate the `echo` command and the specific LEMC verb.

**Example (shell script helper functions):**
'''shell
#!/bin/sh

# Helper function to truncate HTML
lemc_html_trunc() {
  echo "lemc.html.trunc;$1"
}

# Helper function to append CSS
lemc_css_append() {
  echo "lemc.css.append;$1"
}

# Helper function to execute JS (by truncating and setting new JS)
lemc_js_exec() {
  echo "lemc.js.exec;$1"
}

# Helper function to set an environment variable
lemc_set_env() {
  echo "lemc.env;$1"
}

# --- Usage ---
lemc_html_trunc "<h2>Updated Content</h2><p>This is the new HTML.</p>"
lemc_css_append ".important { font-weight: bold; color: red; }"
lemc_js_exec "document.body.style.filter = 'invert(1)';"
lemc_set_env "LAST_UPDATE=$(date)"

'''

### Benefits of Helper Functions:

*   **Readability**: Makes the main part of your script easier to understand.
*   **Maintainability**: If the LEMC verb syntax changes, you only need to update the helper functions.
*   **Reduced Errors**: Less chance of typos in the command prefixes.
*   **Abstraction**: The helper function handles the "how" (the `echo` and prefix), so your main script focuses on "what" (the content).

When a command is sent using these verbs (either directly or via a helper function), the `yeschef` application will:
1.  **Log the raw message**: The original string (e.g., `lemc.html.trunc;<h1>Title</h1>`) is logged by `yeschef`.
2.  **Process the command**: The appropriate action (truncating CSS, appending HTML, etc.) is taken, and the content is streamed to the connected client(s).

This provides a mechanism for both logging the container activity and updating the user interface in real-time. 
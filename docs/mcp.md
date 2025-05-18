# Using LEMC's MCP Service

This document provides a minimal example of how to enable the **Model Context Protocol (MCP)** feature in Let'em Cook and interact with it from an external client. MCP exposes your app's recipes and pages via JSON‑RPC over HTTP and streams output using **Server‑Sent Events (SSE)**.

## 1. Enable MCP for an App

1. Create or open an app in the LEMC web UI.
2. Toggle **MCP Enabled** for that app. When enabled, the app generates an **API key** that clients must supply when connecting.
3. Copy the API key from the **MCP KEY** modal on the app page.

## 2. Connect to the MCP Server

Each app exposes two endpoints:

* `GET /mcp/app/<UUID>` – opens an SSE stream. Include the API key using the `X-API-Key` header.
* `POST /mcp/app/<UUID>` – send MCP JSON‑RPC requests. Use the same `X-API-Key` header.

Example shell session using `curl`:

```bash
APP_UUID=<your-app-uuid>
API_KEY=<your-api-key>

# Start the SSE stream in one terminal
curl -H "X-API-Key: $API_KEY" http://localhost:5362/mcp/app/$APP_UUID
```

In a second terminal you can issue JSON‑RPC commands:

```bash
# List available pages
curl -X POST -H "X-API-Key: $API_KEY" \
     -d '{"jsonrpc":"2.0","id":1,"method":"lemc.pages"}' \
     http://localhost:5362/mcp/app/$APP_UUID

# List all recipes
curl -X POST -H "X-API-Key: $API_KEY" \
     -d '{"jsonrpc":"2.0","id":2,"method":"lemc.recipes"}' \
     http://localhost:5362/mcp/app/$APP_UUID
```

Responses will be streamed to the SSE connection as JSON‑RPC result objects.

## 3. Running a Recipe

MCP exposes a single tool named `run-recipe`. It allows you to execute any recipe defined for the app.

```bash
# Call the run-recipe tool
curl -X POST -H "X-API-Key: $API_KEY" \
     -d '{"jsonrpc":"2.0","id":3,"method":"tools/call",
          "params":{"name":"run-recipe",
                   "arguments":{"page":1,"recipe":"example"}}}' \
     http://localhost:5362/mcp/app/$APP_UUID
```

The command triggers the recipe just as if it were run from the web UI. Status messages such as `--MCP JOB STARTED--` and `--MCP JOB FINISHED--` appear on the SSE stream followed by any output produced by the recipe steps.

## 4. Listing and Reading Resources

Recipes may contain page wiki content that can be fetched via `resources/list` and `resources/read`:

```bash
# List all resources (e.g., page wikis)
curl -X POST -H "X-API-Key: $API_KEY" \
     -d '{"jsonrpc":"2.0","id":4,"method":"resources/list"}' \
     http://localhost:5362/mcp/app/$APP_UUID

# Read a specific resource
curl -X POST -H "X-API-Key: $API_KEY" \
     -d '{"jsonrpc":"2.0","id":5,"method":"resources/read",
          "params":{"uri":"lemc://app/<UUID>/wiki/1"}}' \
     http://localhost:5362/mcp/app/$APP_UUID
```

The returned `contents` array includes the wiki text or other resource data.

## Summary

Once MCP is enabled for an app and you have its API key, you can programmatically inspect and execute recipes using simple HTTP requests. The SSE stream delivers real‑time feedback while JSON‑RPC calls allow tooling or other agents to drive Let'em Cook workflows.

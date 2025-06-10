# lemcli

`lemcli` is a minimal command-line tool for creating cookbooks and managing secrets outside your cookbook directory.

## Creating a Cookbook

Initialize a new cookbook directory. `lemcli` will automatically prefix the directory name with `lemc-` if it doesn't already start with that prefix. The directory will contain a starter `cookbook.yaml` file.

```
lemcli cookbook init my-example-app
```

This command creates a folder `lemc-my-example-app/` with `cookbook.yaml` inside.

## Storing Secrets

Secrets for a cookbook can be stored in `~/.lemc/secrets/<cookbook>`.

```
lemcli secrets init my-cookbook
lemcli secrets set my-cookbook API_KEY -value mysecret
lemcli secrets get my-cookbook API_KEY
```

`lemcli` ensures the secrets directory is created with permission `0700`. Each secret is written to a file named after the key with mode `0600`.

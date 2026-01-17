---
title: Getting started
description: Get started by setting up Toki and minimal config
---

Welcome to Toki! This tutorial will guide you through the essentials of installing Toki and creating your first pipeline config. By the end of this tutorial, you'll have a working Toki pipeline with your custom command.

## Prerequisites

Before we begin make sure you have a terminal or command prompt.

## Step 1: Downloading Toki binary

First, let's download Toki binary file for your OS from [releases page](https://github.com/pratyushtiwary/toki/releases)

:::tip[Info]
You can add the downloaded binary in your system's PATH environment variable to access `toki` from any directory
:::

## Step 2: Writing your first pipeline config

Toki is driven via config(s). We'll start by creating a simple pipeline config. Copy and paste the following config into `pipeline-config.json` file:
```json
[
  {
    "command": "sleep 120",
    "auth": [
      {
        "strategy": "custom",
        "params": {
          "expiry": 2,
          "refresh_after": 1,
          "command": "sleep 1 && echo 1"
        }
      }
    ]
  }
]

```

The above config contains a single step with `sleep 120` as the main command and a custom auth strategy with refresh rate set to 1 minute interval.

To run this config run the following command in your terminal: `toki pipeline-config.json`, this would spawn a child process which will run the main command in the background while the main process takes care of handling auth expiry logic.

🚀 Congratulations you just ran your first Toki pipeline!

:::tip[Did you know?]
There are 3 types of config supported in Toki, you can learn more about each config from here
:::

## Next steps

Now that you have a simple pipeline config, you can:
- Try adding more steps,
- Learn about different config types,
- Explore examples

Happy orchestrating with Toki!
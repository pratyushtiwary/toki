---
title: Chain commands
description: This guide shows how you can chain multiple commands into a single command in Toki config
---

This guide shows how you can chain multiple commands into a single command in Toki config

## Prerequisites

Before we start, you should have:
- Copy of Toki binary in your system,
- A shell environment

## Simple use-case

Toki runs all the command via `/bin/sh` on Unix machines or `cmd.exe` on Windows, this allows commands to be chainable. So you can add commands such as `sleep 5 && sleep 2` and it would work just find with Toki.

Here's an example pipeline config:
```json
[
  {
    "command": "sleep 5m && sleep 5m",
    "auth": [
      {
        "strategy": "custom",
        "params": {
          "expiry": 5,
          "refresh_after": 2,
          "command": "sleep 2s && sleep 3s"
        }
      }
    ]
  }
]
```

## Advanced use-case

Toki also supports using pipes and output redirection. Here's an example of using output redirection in Toki pipeline config:
```json
[
  {
    "command": "sleep 20s && echo '12345678\n910111213141516' | grep -oh -E '8|16' > test.txt",
    "auth": [
      {
        "strategy": "custom",
        "params": {
          "expiry": 10,
          "refresh_after": 5,
          "command": "sleep 5s"
        }
      }
    ]
  }
]
```

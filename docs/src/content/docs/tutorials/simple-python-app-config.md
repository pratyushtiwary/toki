---
title: Configuring a simple python application
description: This guide goes over how you can configure a simple python application
---

Have you ever wrote a python script that takes more than an hour and your auth credentials are only valid for an hour? It's frustrating when your code fails because of token expired or unauthorized exception 🥲. There are workaround such as wrapping your code in a try/catch block and running auth command when catch block hits for token expired or unauthorized case but most of these workarounds are hacky and require changes in code.

Toki tries to solve this problem by suspending the process of your application and running auth commands as they are about to expire. Toki achieves this by the use of syscalls.

## The Goal

This tutorial will guide us through how we can create a Toki pipeline for a Python script. By the end of this tutorial you will have good understanding of internals of Toki to some extent.

## Prerequisites

Before we start, you should have:
- Copy of Toki binary in your system,
- A shell environment

## Step 1: Laying the foundation

Before we start we need an auth provider and a dummy python file. We'll be using https://dummyjson.com/docs/auth as the auth provider for this tutorial.

Let's start by writing a dummy authenticator CLI in Python that'll save the token in a file under userspace:
```python
#!/usr/bin/env python3
import requests

auth_url = "https://dummyjson.com/auth/login"
try:
    response = requests.post(
        auth_url,
        json={"username": "emilys", "password": "emilyspass", "expiresInMins": 60},
    )

    assert response.status_code == 200

    data = response.json()

    with open(".tutorial-dummy-creds", "w") as fd:
        fd.write(data['accessToken'])

    print("Access token refreshed successfully!")
except Exception as e:
    print(f"Failed to refresh access token, error: {e}")
```

Save the above code in a file called `test-refresh.py`. In the code above we are fetching the JWT token and saving it in a file called `.tutorial-dummy-creds`, the JWT token will expire in 60 mins to simulate real-world scenario

Once we have the refresh file saved, let's create a client which would be using this saved JWT token to hit secure endpoints:
```python
#!/usr/bin/env python3
import requests
from time import sleep

profile_url = "https://dummyjson.com/auth/me"

with open("./.tutorial-dummy-creds", "r") as fd:
    auth_token = fd.read()

for i in range(70):
    response = requests.get(profile_url, headers={"Authorization": f"Bearer {auth_token}"})
    assert response.status_code == 200
    data = response.json()
    assert data["role"] == "admin"
    sleep(60)
```

Save the above code in a file called `test-client.py`, the code above is using the auth token we saved to make request to a secure endpoint, the code will fail if we the token gets expired in between.

:::note
Client code will take approximately 70 mins to finish its execution, whereas the auth token is only valid for 60 mins
:::

## Step 2: Defining Configuration

Now that we have the python files ready, let's start by creating a Toki pipeline config so that we can run the above file with Toki.

Start by creating an empty file in the same directory as the python files named `test.pipeline.json`.

Open the file and copy-paste the following content in it:
```json
[
  {
    "command": "python3 test-client.py",
    "auth": [
      {
        "strategy": "custom",
        "params": {
          "expiry": 60,
          "refresh_after": 50,
          "command": "python3 test-refresh.py"
        }
      }
    ]
  }
]
```

In the above config file we've added 1 step which contains our main command `python3 test-client.py` and auth refresh command `python3 test-refresh.py`. Notice how we are also defining a strategy and its relevant params, strategies are way to tell Toki how it should compute the time for token expiry, you can have multiple auth refresh configs for a single main command.

:::tip
If you were to use multiple auth refresh configs then Toki will run all of them when anyone of them expires.
:::

## Step 3: Seeing it all work

Now that we have to python files and config ready we can test it by running `toki test.pipeline.json`, this should ideally start Toki and the output should look similar to the one below:
```text
 ______   ______     __  __     __    
/\__  _\ /\  __ \   /\ \/ /    /\ \   
\/_/\ \/ \ \ \/\ \  \ \  _"-.  \ \ \  
   \ \_\  \ \_____\  \ \_\ \_\  \ \_\ 
    \/_/   \/_____/   \/_/\/_/   \/_/ 

Parsing file: test.pipeline.json
1 step(s) loaded, executing step(s) sequentially

----------------- Executing step 1 -------------------
Starting auth refresh before running `python3 test-client.py` command
Running: `python3 test-refresh.py` for auth refresh
Done with initial auth refresh, executing `python3 test-client.py`
```

## Conclusion

You've now implemented a pipeline config which makes refreshing auth easier without making any changes in the source code 🚀!

By using Toki to orchestrate auth, you have:
- A **single source of truth** for how your script is orchestrated,
- A **clean separation** between your application code and orchestration configuration

For even larger scripts, consider moving the step we've created in this tutorial in a project config so you can reuse it across different pipeline config's steps, this would make it more maintainable.
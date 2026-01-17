---
title: Handling Toki events
description: This guide goes over how your program can listen to Toki's events
---

Processess spawned by Toki can listen to event broadcasted by Toki. These events can help perform cleanups or additional tasks. As of writing this guide there's only 1 event that is broadcasted by Toki.

Events:
- `AUTH_REFRESHED`: This event is broadcasted by Toki once all auth commands are done running and Toki resumed the main process

## Prerequisites

Before we start, you should have:
- Cloned [Toki repo](https://github.com/pratyushtiwary/toki),
- A shell environment

## Step 1: Understanding the architecture

Toki starts a server which child processess can connect to. The access to server is by default restricted and only client with correct token are allowed to connect. Toki injects special env variables to the spawned child processess. Following is the description of env vars injected by Toki:
- `TOKI_PORT`: This env var can be used by child process to connect to Toki's server on correct port instead of hardcoding the default port(i.e. `3110`).
- `TOKI_TOKEN`: Each client has to provide a passkey after connection, if passkey is invalid connection is closed by the server.

:::tip
You can disable restricted access behaviour by setting `dev` flag while running Toki, this flag disables the token check logic and is helpful when debugging Toki clients
:::

## Step 2: Writing Toki client

In this tutorial we'll be writing a Toki client in Python, but you can use any other language since the concept is same only syntax would differ.

Copy paste the below code snippet in `client.py` file:
```python
#!/usr/bin/env python3
import os
import asyncio
from typing import Callable, Awaitable, Optional, List
from threading import Thread
from queue import Queue
from enum import Enum

class TokiEvent(Enum):
    AUTH_REFRESHED = "AUTH_REFRESHED"

class TokiClient:

    def __init__(self, timeout: int = 5, host: str = "localhost", port: int = None, passkey: str = None):
        if port is None:
            port = int(os.environ.get("TOKI_PORT", 3110))

        if passkey is None:
            passkey = os.environ.get("TOKI_TOKEN")

        if passkey is None:
            raise ValueError("Invalid `passkey` is provided")

        self.__host = host
        self.__port = port
        self.__passkey = passkey
        self.timeout = timeout
        self.reader: Optional[asyncio.StreamReader] = None
        self.writer: Optional[asyncio.StreamWriter] = None
        self.handlers: List[Callable[[str], Awaitable[None]]] = []
        self.queue = Queue()

        self.loop: Optional[asyncio.AbstractEventLoop] = None
        self.thread: Optional[Thread] = None

        self._start_background_loop()

    @property
    def host(self):
        return self.__host

    @property
    def port(self):
        return self.__port

    @property
    def passkey(self):
        return self.__passkey

    def _start_background_loop(self):
        self.thread = Thread(
            target=self._run_event_loop, daemon=True, name="TokiEventLoop"
        )
        self.thread.start()

    def _run_event_loop(self):
        self.loop = asyncio.new_event_loop()
        asyncio.set_event_loop(self.loop)

        try:
            self.loop.run_until_complete(self._async_main())
        except Exception as e:
            raise ValueError(f"Failed to run main loop inside spawned thread, error: {e}")
        finally:
            self.loop.close()

    async def _async_main(self):
        if not await self._connect():
            return

        await self._listen()

    async def _connect(self):
        try:
            self.reader, self.writer = await asyncio.wait_for(
                asyncio.open_connection(self.host, self.port), timeout=self.timeout
            )

            self.writer.write(self.passkey.encode())
            await self.writer.drain()
            return True

        except Exception as e:
            raise ValueError(f"Failed to connect to Toki server, error: {e}")

    async def _handle_event(self, event_data: str):
        event_data = event_data.strip()
        event_data_enum: TokiEvent = TokiEvent[event_data]
        for handler in self.handlers:
            await handler(event_data_enum)

    async def _listen(self):
        while True:
            try:
                msg = await self.reader.read(4096)

                if not msg:
                    break

                await self._handle_event(msg.decode('utf-8'))
            except:
                break

    def on_event(self, handler: Callable[[str], Awaitable[None]]):
        if not asyncio.iscoroutinefunction(handler):
            raise ValueError("Attached handler should be an async function")

        self.handlers.append(handler)

    def event_handler(self, func: Callable[[str], Awaitable[None]]):
        if not asyncio.iscoroutinefunction(func):
            raise ValueError("Attached handler should be an async function")

        self.handlers.append(func)

        return func
    
    def __enter__(self):
        return self
    
    def __exit__(self):
        self.__del__()

    def __del__(self):
        if self.writer:
            self.writer.close()
```

This code above spawn a separate thread in the background which starts the connection with Toki server and actively listens to events published by it. Once an event is captured the code notifies the subscribers by using a decorator exposed by the client.

## Step 3: Using Toki client

Start by creating a new file called `test.py` and copy paste the code below in that file:
```python
#!/usr/bin/env python3
import time
from client import TokiClient, TokiEvent

client = TokiClient()

@client.event_handler
async def handle_event(event_data: TokiEvent):
    if event_data == TokiEvent.AUTH_REFRESHED:
        print("Auth refreshed, refreshing client...")


for i in range(10):

    print(i + 1)

    time.sleep(1)
```

The code above registers `handle_event` function which is invoked by the client when it receives any event from Toki. Notice how an additional check is added on the consumer side to invoke custom logic for specific event type(`AUTH_REFRESHED` in this case), this is done to promote separation of concern.


Congratulations on writing your first Toki client! 🚀
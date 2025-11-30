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

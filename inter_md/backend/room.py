import abc
import aiohttp.web
import collections
import asyncio
import typing


class RoomStateMachine:

    def __init__(self):
        self.teared_down = asyncio.Event()
        self.instantiated = asyncio.Event()
        self.teared_down.set()

    def is_teared_down(self) -> bool:
        return self.teared_down.is_set()

    async def wait_for_tear_down(self):
        await self.teared_down.wait()

    def set_teared_down(self):
        self.teared_down.set()

    def clear_teared_down(self):
        self.teared_down.clear()

    def is_instantiated(self) -> bool:
        return self.instantiated.is_set()

    async def wait_for_instantiate(self):
        await self.instantiated.wait()

    def set_instantiated(self):
        self.instantiated.set()

    def clear_instantiated(self):
        self.instantiated.clear()


class Room(abc.ABC):

    def __init__(self, name: str, send_message: collections.abc.Coroutine):
        super().__init__()
        self.name = name
        self.send_message = send_message
        self.attached_websockets: typing.List[aiohttp.web.WebSocketResponse] = [
        ]
        self.state = RoomStateMachine()

    def __len__(self) -> int:
        return len(self.attached_websockets)

    @abc.abstractmethod
    async def instantiate(self):
        raise NotImplementedError

    @abc.abstractmethod
    async def on_message(self, message: dict):
        raise NotImplementedError

    @abc.abstractmethod
    async def tear_down(self):
        raise NotImplementedError

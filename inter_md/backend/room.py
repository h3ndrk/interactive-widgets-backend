import abc
import aiohttp.web
import collections
import asyncio
import typing

from .room_state_machine import RoomStateMachine


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
    async def handle_message(self, message: dict):
        raise NotImplementedError

    @abc.abstractmethod
    async def tear_down(self):
        raise NotImplementedError

import abc
import aiohttp.web
import collections
import asyncio
import logging
import typing

from .. import contexts
from .. import rooms


class Room(abc.ABC):

    def __init__(self, context: contexts.Context, configuration: dict, name: str, send_message: collections.abc.Coroutine):
        super().__init__()
        self.context = context
        self.configuration = configuration
        self.name = name
        self.send_message = send_message
        self.logger = logging.getLogger(self.configuration['logger_name_room'])
        self.attached_websockets: typing.List[aiohttp.web.WebSocketResponse] = [
        ]
        self.state = rooms.RoomStateMachine()

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

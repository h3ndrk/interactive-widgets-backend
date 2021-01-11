import abc
import aiohttp.web
import collections
import asyncio
import logging
import typing

import interactive_widgets.backend.contexts.context


class Room(metaclass=abc.ABCMeta):

    def __init__(self, context: interactive_widgets.backend.contexts.context.Context, configuration: dict, name: str, send_message: collections.abc.Coroutine):
        super().__init__()
        self.context = context
        self.configuration = configuration
        self.name = name
        self.send_message = send_message
        self.logger = logging.getLogger(self.configuration['logger_name_room'])
        self.attached_websockets: typing.List[aiohttp.web.WebSocketResponse] = [
        ]

    def __len__(self) -> int:
        return len(self.attached_websockets)

    def attach_websocket(self, websocket: aiohttp.web.WebSocketResponse):
        self.attached_websockets.append(websocket)

    def detach_websocket(self, websocket: aiohttp.web.WebSocketResponse):
        self.attached_websockets.remove(websocket)

    def is_emtpy(self) -> bool:
        return len(self.attached_websockets) == 0

    @abc.abstractmethod
    async def handle_message(self, message: dict):
        raise NotImplementedError

    @abc.abstractmethod
    async def update(self):
        raise NotImplementedError

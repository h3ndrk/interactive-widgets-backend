import aiodocker
import aiohttp.web
import asyncio
import logging
import typing

import interactive_widgets.backend.shield
import interactive_widgets.backend.contexts.context
import interactive_widgets.backend.rooms.get
import interactive_widgets.backend.rooms.room


# Test Cases (xC = WebSocket x connected, xD = WebSocket x disconnected, IN = Instantiated, TD = Torn-Down):
# - 1C, IN, 1D, TD
# - 1C, 1D, IN, TD
# - 1C, IN, 2C, 2D, 1D, TD
# - 1C, 2C, IN, 2D, 1D, TD
# - 1C, 2C, 2D, IN, 1D, TD
# - 1C, IN, 2C, 1D, 2D, TD
# - 1C, 2C, IN, 1D, 2D, TD
# - 1C, IN, 1D, 2C, 2D, TD
# - 1C, 1D, IN, 2C, 2D, TD
# - 1C, 2C, 1D, IN, 2D, TD
# - 1C, 1D, 2C, IN, 2D, TD
# - 1C, 2C, 2D, 1D, IN, TD
# - 1C, 2C, 1D, 2D, IN, TD
# - 1C, 1D, 2C, 2D, IN, TD
# - 1C, IN, 1D, TD, 2C, IN, 2D, TD
# - 1C, 1D, IN, TD, 2C, IN, 2D, TD
# - 1C, IN, 1D, 2C, TD, IN, 2D, TD
# - 1C, 1D, IN, 2C, TD, IN, 2D, TD
# - 1C, IN, 1D, TD, 2C, 2D, IN, TD
# - 1C, 1D, IN, TD, 2C, 2D, IN, TD


class RoomConnection:

    def __init__(self, context: interactive_widgets.backend.contexts.context.Context, configuration: dict, rooms: typing.Dict[str, interactive_widgets.backend.rooms.room.Room], room_name: str, websocket: aiohttp.web.WebSocketResponse):
        self.context = context
        self.configuration = configuration
        self.rooms = rooms
        self.room_name = room_name
        self.websocket = websocket
        self.logger = logging.getLogger(
            self.configuration['logger_name_room_connection'])

    async def __aenter__(self):
        # The context manager in this class manages the instantiation and tear down of rooms.
        # It has been designed carefully s.t. no asynchronous race conditions can happen.
        # The first part ensures that a constructed (not necessarily instantiated) room exists
        # and is registered in the page's room dictionary (self.rooms). The new websocket of
        # this room connection is added to the room, resulting in either 1 websocket for new
        # rooms or more than 1 websockets for existing rooms. Since no asynchronous context
        # switch can happen, the room management and websocket attachment is asynchronously
        # atomic.
        try:
            room = self.rooms[self.room_name]
            self.logger.debug(f'Using existing room {self.room_name}...')
        except KeyError:
            self.logger.debug(f'Creating room {self.room_name}...')
            room = interactive_widgets.backend.rooms.get.get(
                self.configuration['type'],
            )(
                self.context,
                self.configuration,
                self.room_name,
                self._send_message,
            )
            self.rooms[self.room_name] = room

        self.logger.debug(f'Attaching websocket {id(self.websocket)}...')
        room.attach_websocket(self.websocket)
        try:
            await room.update()
        except:
            await interactive_widgets.backend.shield.shield(self._detach(room))
            raise

        return room

    async def __aexit__(self, *args, **kwargs):
        await interactive_widgets.backend.shield.shield(self._detach(self.rooms[self.room_name]))

    async def _detach(self, room: interactive_widgets.backend.rooms.room.Room):
        self.logger.debug(f'Detaching websocket {id(self.websocket)}...')
        room.detach_websocket(self.websocket)
        try:
            await room.update()
        finally:
            if room in self.rooms.values() and room.is_emtpy():
                self.logger.debug('Deleting room...')
                del self.rooms[self.room_name]

    async def _send_message(self, message):
        self.logger.debug('Sending message to all attached websockets...')
        for websocket in self.rooms[self.room_name].attached_websockets:
            self.logger.debug(
                f'Sending message to websocket {id(websocket)}...')
            await websocket.send_json(message)

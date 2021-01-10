import aiodocker
import aiohttp.web
import logging
import typing

import interactive_widgets.backend.contexts.context
import interactive_widgets.backend.rooms.get
import interactive_widgets.backend.rooms.room


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
        room.attached_websockets.append(self.websocket)

        # This code instantiates a room or waits for already running instantiation. The
        # expected exceptions are asyncio.CancelledError and all exceptions raised during
        # instantiation (failures). In all cases the partially instantiated room must be
        # torn down.
        try:
            if len(room.attached_websockets) == 1:
                self.logger.debug('First attached websocket, instantiating...')
                await room.instantiate()
            else:
                self.logger.debug('Waiting for instantiation...')
                await room.state.wait_for_instantiate()
        except:
            await self._tear_down(force_if_not_instantiated=True)
            raise

        return room

    async def __aexit__(self, *args, **kwargs):
        await self._tear_down(force_if_not_instantiated=False)

    async def _tear_down(self, force_if_not_instantiated: bool):
        # Same as during instantiation, care must be taken if exceptions are raised during
        # tear down (e.g. asyncio.CancelledError or failure exceptions). In all cases the
        # room must be cleaned up correctly.
        room = self.rooms[self.room_name]

        self.logger.debug(f'Detaching websocket {id(self.websocket)}...')
        room.attached_websockets.remove(self.websocket)

        try:
            if len(room.attached_websockets) == 0 and (force_if_not_instantiated or room.state.is_instantiated()):
                self.logger.debug('Last websocket detached, tearing down...')
                room.state.clear_instantiated()
                await room.tear_down()
        finally:
            if len(room.attached_websockets) == 0:
                self.logger.debug('Deleting room...')
                del self.rooms[self.room_name]

    async def _send_message(self, message):
        self.logger.debug('Sending message to all attached websockets...')
        for websocket in self.rooms[self.room_name].attached_websockets:
            self.logger.debug(
                f'Sending message to websocket {id(websocket)}...')
            await websocket.send_json(message)

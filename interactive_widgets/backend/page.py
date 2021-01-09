import aiodocker
import aiohttp.web
import json
import logging
import pathlib
import typing

import interactive_widgets.backend.contexts.context
import interactive_widgets.backend.rooms.room
import interactive_widgets.backend.rooms.room_connection


class Page:

    def __init__(self, context: interactive_widgets.backend.contexts.context.Context, configuration: dict, url: pathlib.PurePosixPath, application: aiohttp.web.Application):
        self.context = context
        self.configuration = configuration
        self.url = url
        self.application = application
        self.logger = logging.getLogger(self.configuration['logger_name_page'])
        self.application.add_routes([
            aiohttp.web.get(str(self.url / 'ws'), self._handle_websocket),
        ])
        self.rooms: typing.Dict[str,
                                interactive_widgets.backend.rooms.room.Room] = {}

    def _connect(self, room_name: str, websocket: aiohttp.web.WebSocketResponse):
        return interactive_widgets.backend.rooms.room_connection.RoomConnection(
            self.context,
            self.configuration,
            self.rooms,
            room_name,
            websocket,
        )

    async def _handle_websocket(self, request: aiohttp.web.Request):
        try:
            room_name = request.query['roomName']
            self.logger.debug(f'Extracted room name: {room_name}')
        except KeyError:
            raise aiohttp.web.HTTPBadRequest(reason='Missing roomName')

        websocket = aiohttp.web.WebSocketResponse(heartbeat=10)
        await websocket.prepare(request)
        self.logger.info(
            f'Got websocket {id(websocket)} from {request.remote}')

        async with self._connect(room_name, websocket) as room:
            while True:
                message = await websocket.receive()
                if message.type == aiohttp.web.WSMsgType.TEXT:
                    parsed_message = json.loads(message.data)
                    self.logger.debug(parsed_message)
                    await room.handle_message(parsed_message)
                else:
                    self.logger.warning(f'Unexpected message: {message}')
                    break

        return websocket

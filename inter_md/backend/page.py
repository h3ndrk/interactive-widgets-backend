import aiodocker
import aiohttp.web
import json
import logging
import typing

import inter_md.backend.contexts.context
import inter_md.backend.rooms.room
import inter_md.backend.rooms.room_connection


class Page:

    def __init__(self, context: inter_md.backend.contexts.context.Context, configuration: dict):
        self.context = context
        self.configuration = configuration
        self.logger = logging.getLogger(self.configuration['logger_name_page'])
        self.application = aiohttp.web.Application()
        self.application.add_routes([
            # aiohttp.web.get('/index', self._handle_index),
            # aiohttp.web.get('/red-ball.png', self._handle_red_ball),
            aiohttp.web.get('/ws', self._handle_websocket),
        ])
        self.rooms: typing.Dict[str, inter_md.backend.rooms.room.Room] = {}

    def _connect(self, room_name: str, websocket: aiohttp.web.WebSocketResponse):
        return inter_md.backend.rooms.room_connection.RoomConnection(
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

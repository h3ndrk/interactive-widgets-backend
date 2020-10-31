import aiodocker
import aiohttp.web
import json
import typing

from . import contexts
from . import rooms


class Page:

    def __init__(self, context: contexts.Context, configuration: dict):
        self.context = context
        self.configuration = configuration
        self.application = aiohttp.web.Application()
        self.application.add_routes([
            # aiohttp.web.get('/index', self._handle_index),
            # aiohttp.web.get('/red-ball.png', self._handle_red_ball),
            aiohttp.web.get('/', self._handle_websocket),
        ])
        self.rooms: typing.Dict[str, rooms.Room] = {}

    def connect(self, room_name: str, websocket: aiohttp.web.WebSocketResponse):
        return rooms.RoomConnection(
            self.context,
            self.configuration,
            self.rooms,
            room_name,
            websocket,
        )

    async def _handle_websocket(self, request: aiohttp.web.Request):
        try:
            room_name = request.query['roomName']
            print(f'Extracted room name: {room_name}')
        except KeyError:
            raise aiohttp.web.HTTPBadRequest(reason='Missing roomName')

        websocket = aiohttp.web.WebSocketResponse(heartbeat=10)
        await websocket.prepare(request)
        print(f'Got websocket {id(websocket)} from {request.remote}')

        async with self.connect(room_name, websocket) as room:
            while True:
                message = await websocket.receive()
                if message.type == aiohttp.web.WSMsgType.TEXT:
                    parsed_message = json.loads(message.data)
                    print(parsed_message)
                    await room.handle_message(parsed_message)
                else:
                    print(f'Unexpected message: {message}')
                    break

        return websocket

import aiodocker
import aiohttp.web
import typing

from .room import Room
from .room_connection import RoomConnection


class Page:

    def __init__(self, docker: aiodocker.Docker):
        self.docker = docker
        self.rooms: typing.Dict[str, Room] = {}

    def connect(self, room_name: str, websocket: aiohttp.web.WebSocketResponse):
        return RoomConnection(self.docker, self.rooms, room_name, websocket)

    async def _handle_websocket(self, request: aiohttp.web.Request):
        # extract room name
        try:
            room_name = request.query['roomName']
        except KeyError:
            raise aiohttp.web.HTTPBadRequest(reason='Missing roomName')

        # initiate websocket
        websocket = aiohttp.web.WebSocketResponse()
        await websocket.prepare(request)
        print(f'Got websocket {id(websocket)} from {request.remote}')

        async with self.connect(room_name, websocket) as room:
            while True:
                message = await websocket.receive_json()
                print(message)
                await room.handle_message(message)

        return websocket

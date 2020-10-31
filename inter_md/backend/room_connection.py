import aiodocker
import aiohttp.web
import typing

from .room import Room
from .docker_room import DockerRoom


class RoomConnection:
    def __init__(self, docker: aiodocker.Docker, rooms: typing.Dict[str, Room], room_name: str, websocket: aiohttp.web.WebSocketResponse):
        # TODO: make docker member a static member of DockerRoom
        self.docker = docker
        self.rooms = rooms
        self.room_name = room_name
        self.websocket = websocket

    async def __aenter__(self):
        try:
            print(f'Using existing room {self.room_name}...')
            room = self.rooms[self.room_name]
        except KeyError:
            print(f'Creating room {self.room_name}...')
            # TODO: factor for returning new rooms with different types
            room = DockerRoom(self.docker, self.room_name, self._send_message)
            self.rooms[self.room_name] = room

        print(f'Attaching websocket {id(self.websocket)}...')
        room.attached_websockets.append(self.websocket)

        if len(room.attached_websockets) == 1:
            print('First attached websocket, instantiating...')
            await room.instantiate()
        else:
            print('Waiting for instantiation...')
            await room.state.wait_for_instantiate()

        return room

    async def __aexit__(self, *args, **kwargs):
        room = self.rooms[self.room_name]

        print(f'Detaching websocket {id(self.websocket)}...')
        room.attached_websockets.remove(self.websocket)

        if len(room.attached_websockets) == 0 and room.state.is_instantiated():
            print('Last websocket detached, tearing down...')
            room.state.clear_instantiated()
            await room.tear_down()
            if len(room.attached_websockets) == 0:
                del self.rooms[self.room_name]

    async def _send_message(self, message):
        print('Sending message to all attached websockets...')
        for websocket in self.rooms[self.room_name].attached_websockets:
            print(f'Sending message to websocket {id(websocket)}...')
            await websocket.send_json(message)
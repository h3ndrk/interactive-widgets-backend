import click
import aiodocker
import aiohttp.web
import asyncio
import logging
import typing

from .room import Room
from .docker_room import DockerRoom

logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s  %(name)-17s  %(levelname)-8s  %(message)s',
)


class MockedWebSocket:

    def __init__(self, messages: list, delay: float):
        self.messages = messages
        self.delay = delay

    def __aiter__(self):
        return self

    async def __anext__(self):
        await asyncio.sleep(self.delay)
        try:
            return self.messages.pop(0)
        except IndexError:
            raise StopAsyncIteration

    async def receive_json(self):
        await asyncio.sleep(self.delay)
        return self.messages.pop(0)
        
    async def send_json(self, message):
        print('Send message:', message)

# class InstantiatedRoom(Room):
#     def __init__(self, ...):
#         super().__init__(...)
#         # super(Room, self).__init__(...)
#         # Room.__init__(self, ...)
#         self.attached_websockets: typing.List[aio...Websocket...]


class RoomConnection:
    def __init__(self, docker: aiodocker.Docker, rooms: typing.Dict[str, Room], room_name: str, websocket: aiohttp.web.WebSocketResponse):
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


class Page:

    def __init__(self, docker: aiodocker.Docker):
        self.docker = docker
        self.rooms: typing.Dict[str, Room] = {}

    def connect(self, room_name: str, websocket):
        return RoomConnection(self.docker, self.rooms, room_name, websocket)

# class Room:
#     pass


async def send_message(message):
    print(message)


async def async_main(**arguments):
    async with aiodocker.Docker() as docker:
        page = Page(docker)
        websocket = MockedWebSocket([
            {
                'executor': 'init',
                'message': 42,
            },
            {
                'executor': 'init',
                'message': 42,
            },
            {
                'executor': 'init',
                'message': 42,
            },
        ], 1)
        websocket2 = MockedWebSocket([
            {
                'executor': 'init',
                'message': 42,
            },
            {
                'executor': 'init',
                'message': 42,
            },
            {
                'executor': 'init',
                'message': 42,
            },
        ], 1)
        async def task_function(websocket):
            async with page.connect('cba', websocket) as room:
                while True:
                    message = await websocket.receive_json()
                    print(message)
                    await room.handle_message(message)
        
        task1 = asyncio.create_task(task_function(websocket))
        task2 = asyncio.create_task(task_function(websocket2))
        await task1
        await task2
        # room = DockerRoom(docker, 'abc', send_message)
        # await room.instantiate()
        # input()
        # await room.tear_down()
        # print('Instantiating...')
        # await room.communicate(websocket)
        # print('Done.')

        # room

        # # RoomConnection
        # async with self.connect(room_name, websocket) as room:
        #     # await connection()
        #     while True:
        #         room.send_message(await websocket.read())

        # room_dict
        # async with room_dict.with_name(room_name) as room:
        #     async with room.connect(websocket) as connection:
        #         await connection()


@click.command()
def main(**arguments):
    asyncio.run(async_main(**arguments))

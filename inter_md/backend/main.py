import click
import aiodocker
import aiohttp.web
import asyncio
import logging
import typing

from .room import Room
from .docker_room import DockerRoom
from .page import Page

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
                    try:
                        message = await websocket.receive_json()
                    except IndexError:
                        break
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

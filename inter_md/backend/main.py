import click
import aiodocker
import asyncio
import logging

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


# class InstantiatedRoom(Room):
#     def __init__(self, ...):
#         super().__init__(...)
#         # super(Room, self).__init__(...)
#         # Room.__init__(self, ...)
#         self.attached_websockets: typing.List[aio...Websocket...]


# class Page:
#     def __init__(self):
#         self.rooms: typing.Dict[str, InstantiatedRoom] = {}
#     def connect(self, name: str, websocket):
#         return RoomConnection(self.rooms, name, websocket)

# class RoomConnection:
#     def __init__(self, rooms, name, websocket):
#         pass
#     async def __aenter__(self):
#         # try:
#         #     room = self.rooms[room_name]
#         #     room.wait_for_instantiation() ...
#         # except:
#         #     room = Room(self.send_message)
#         #     self.rooms[self.name] = room
#         #     await room.instantiate() ...
#         return room
#     async def __aexit__(self, *args, **kwargs):
#         # tear down
#         # finally: ...
#     async def send_message(self, message):
#         for websocket in self.rooms[self.name].attached_websockets:
#             await websocket.send()

# class Room:
#     pass


async def send_message(message):
    print(message)


async def async_main(**arguments):
    async with aiodocker.Docker() as docker:
        room = DockerRoom(docker, 'abc', send_message)
        await room.instantiate()
        input()
        await room.tear_down()
        # websocket = MockedWebSocket([1,2,3], 1)
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

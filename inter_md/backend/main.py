import click
import aiodocker
import asyncio
from .room import Room


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


async def async_main(**arguments):
    async with aiodocker.Docker() as docker:
        room = Room(docker, 'test')
        websocket = MockedWebSocket([1,2,3], 1)
        print('Instantiating...')
        await room.communicate(websocket)
        print('Done.')


@click.command()
def main(**arguments):
    asyncio.run(async_main(**arguments))

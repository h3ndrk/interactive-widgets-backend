import asyncio


class RoomStateMachine:

    def __init__(self):
        self.torn_down = asyncio.Event()
        self.instantiated = asyncio.Event()
        self.torn_down.set()

    def is_torn_down(self) -> bool:
        return self.torn_down.is_set()

    async def wait_for_tear_down(self):
        await self.torn_down.wait()

    def set_torn_down(self):
        self.torn_down.set()

    def clear_torn_down(self):
        self.torn_down.clear()

    def is_instantiated(self) -> bool:
        return self.instantiated.is_set()

    async def wait_for_instantiate(self):
        await self.instantiated.wait()

    def set_instantiated(self):
        self.instantiated.set()

    def clear_instantiated(self):
        self.instantiated.clear()

import aiodocker
import aiohttp.web
import binascii
import asyncio
import logging
import typing


class RoomStateMachine:

    def __init__(self):
        self.teared_down = asyncio.Event()
        self.instantiated = asyncio.Event()
        self.teared_down.set()

    def is_teared_down(self) -> bool:
        return self.teared_down.is_set()

    async def wait_for_tear_down(self):
        await self.teared_down.wait()

    def set_teared_down(self):
        self.teared_down.set()

    def clear_teared_down(self):
        self.teared_down.clear()

    def is_instantiated(self) -> bool:
        return self.instantiated.is_set()

    async def wait_for_instantiate(self):
        await self.instantiated.wait()

    def set_instantiated(self):
        self.instantiated.set()

    def clear_instantiated(self):
        self.instantiated.clear()


class Room:

    def __init__(self, docker: aiodocker.Docker, name: str):
        self.docker = docker
        self.name = name
        self.logger = logging.getLogger('Room')
        self.attached_websockets = []
        self.state = RoomStateMachine()

        self.volume: typing.Optional[aiodocker.docker.DockerVolume] = None

    def __len__(self) -> int:
        return len(self.attached_websockets)

    async def _ensure_instantiated(self):
        if len(self.attached_websockets) == 1:
            self.logger.debug('First attached websocket, instantiating...')
            await self._instantiate()
        else:
            self.logger.debug('Waiting for instantiation...')
            await self.state.wait_for_instantiate()

    async def _instantiate(self):
        self.logger.debug('Waiting for tear down...')
        await self.state.wait_for_tear_down()

        try:
            self.logger.debug('Instantiating...')
            self.state.clear_teared_down()
            self.volume = await self.docker.volumes.create(config={
                'Name': f'inter_md_{binascii.hexlify(self.name.encode("utf-8")).decode("utf-8")}',
                # TODO: labels?
            })
            # TODO: instantiate executors
        except:
            self.logger.error('Failed to instantiate.')
            await self._tear_down()
            raise
        self.logger.info('Instantiated.')
        self.state.set_instantiated()

    async def _ensure_teared_down(self):
        # TODO: is_instantiated check needed?
        if len(self.attached_websockets) == 0 and self.state.is_instantiated():
            self.logger.debug('Last websocket detached, tearing down...')
            self.state.clear_instantiated()
            await self._tear_down()

    async def _tear_down(self):
        try:
            self.logger.debug('Tearing down...')
            self.state.clear_instantiated()
            # TODO: tear down executors
            if self.volume is not None:
                await self.volume.delete()
                self.volume = None
        finally:
            self.logger.info('Teared down.')
            self.state.set_teared_down()

    async def communicate(self, websocket: aiohttp.web.WebSocketResponse):
        self.logger.debug(f'Attaching websocket {id(websocket)}...')
        self.attached_websockets.append(websocket)

        try:
            await self._ensure_instantiated()

            while True:
                message = await websocket.receive_json()
                self.logger.debug(f'Received from {id(websocket)}: {message}')
                # TODO: send message to executors
        finally:
            self.logger.debug(f'Detaching websocket {id(websocket)}...')
            self.attached_websockets.remove(websocket)

            await self._ensure_teared_down()

    async def _on_message_from_executor(self, executor, message):
        self.logger.debug('Sending message to all attached websockets...')
        for websocket in self.attached_websockets:
            self.logger.debug(
                f'Sending message to websocket {id(websocket)}...')
            # TODO:
            # await websocket.send_json({
            #     'widget': self.
            # })

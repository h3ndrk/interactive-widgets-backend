import aiodocker
import binascii
import collections
import logging
import typing

from .. import contexts
from .. import executors
from .. import rooms


class DockerRoom(rooms.Room):

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        assert isinstance(
            self.context,
            contexts.DockerContext,
        )

        self.volume: typing.Optional[aiodocker.docker.DockerVolume] = None

        def wrap_send_message(executor_name: str, send_message: collections.abc.Coroutine):
            async def wrapper(message: typing.Any):
                await send_message({
                    'executor': executor_name,
                    'message': message,
                })
            return wrapper

        self.executors = {
            executor_name: getattr(
                executors,
                executor_configuration['type'],
            )(
                self.context,
                executor_configuration,
                executor_name,
                wrap_send_message(executor_name, self.send_message),
            )
            for executor_name, executor_configuration in self.configuration['executors'].items()
        }

    async def instantiate(self):
        self.logger.debug('Waiting for tear down...')
        await self.state.wait_for_tear_down()

        try:
            self.logger.debug('Instantiating...')
            self.state.clear_torn_down()
            self.volume = await self.context.docker.volumes.create(
                config={
                    'Name': f'inter_md_{binascii.hexlify(self.name.encode("utf-8")).decode("utf-8")}',
                    # TODO: labels?
                },
            )
            for executor_name, executor in self.executors.items():
                self.logger.debug(f'Instantiating executor {executor_name}...')
                await executor.instantiate(self.volume)
        except:
            self.logger.error('Failed to instantiate.')
            import traceback
            traceback.print_exc()
            # TODO: should we tear down here?
            # await self._tear_down()
            raise
        self.logger.info('Instantiated.')
        self.state.set_instantiated()

    async def handle_message(self, message: dict):
        await self.executors[message['executor']].handle_message(message['message'])

    async def tear_down(self):
        try:
            self.logger.debug('Tearing down...')
            self.state.clear_instantiated()
            for executor_name, executor in self.executors.items():
                self.logger.debug(f'Tearing down executor {executor_name}...')
                await executor.tear_down()
            if self.volume is not None:
                await self.volume.delete()
                self.volume = None
        finally:
            self.logger.info('Torn down.')
            self.state.set_torn_down()

import aiodocker
import asyncio
import binascii
import collections
import logging
import traceback
import typing

import interactive_widgets.backend.contexts.docker_context
import interactive_widgets.backend.executors.get
import interactive_widgets.backend.rooms.room


class DockerRoom(interactive_widgets.backend.rooms.room.Room):

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        assert isinstance(
            self.context,
            interactive_widgets.backend.contexts.docker_context.DockerContext,
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
            executor_name: interactive_widgets.backend.executors.get.get(
                f'docker.{executor_configuration["type"]}',
            )(
                self.context,
                executor_configuration,
                executor_name,
                wrap_send_message(executor_name, self.send_message),
            )
            for executor_name, executor_configuration in self.configuration['executors'].items()
        }

    async def instantiate(self):
        # If an exception is raised within this function, this room will be torn down afterwards
        # (no need to revert instantiation here).
        self.logger.debug('Waiting for tear down...')
        await self.state.wait_for_tear_down()

        self.logger.debug('Instantiating...')
        self.state.clear_torn_down()
        self.volume = await self.context.docker.volumes.create(
            config={
                'Name': f'interactive_widgets_{binascii.hexlify(self.name.encode("utf-8")).decode("utf-8")}',
                # TODO: labels?
            },
        )
        for executor_name, executor in self.executors.items():
            self.logger.debug(f'Instantiating executor {executor_name}...')
            await executor.instantiate(self.volume)
        self.logger.info('Instantiated.')
        self.state.set_instantiated()

    async def handle_message(self, message: dict):
        await self.executors[message['executor']].handle_message(message['message'])

    async def tear_down(self):
        self.logger.debug('Tearing down...')
        self.state.clear_instantiated()
        try:
            try:
                self.logger.debug(f'Tearing down executors {", ".join(self.executors.keys())}...')
                results = await asyncio.gather(
                    *(
                        executor.tear_down()
                        for executor in self.executors.values()
                    ),
                    return_exceptions=True,
                )
                exceptions = [result for result in results if result is not None]
                if len(exceptions) > 0:
                    for exception in exceptions:
                        traceback.print_exception(
                            type(exception),
                            exception,
                            exception.__traceback__,
                        )
                    raise RuntimeError('Failed to tear down room\'s executors')
            finally:
                if self.volume is not None:
                    await self.volume.delete()
                    self.volume = None
        finally:
            self.logger.info('Torn down.')
            self.state.set_torn_down()

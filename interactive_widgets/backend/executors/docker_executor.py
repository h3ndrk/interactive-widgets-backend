import abc
import aiodocker
import asyncio
import base64
import binascii
import collections
import logging
import typing

import interactive_widgets.backend.shield
import interactive_widgets.backend.contexts.context
import interactive_widgets.backend.contexts.docker_context


class DockerExecutor(metaclass=abc.ABCMeta):

    def __init__(self, context: interactive_widgets.backend.contexts.context.Context, configuration: dict, name: str, send_message: collections.abc.Coroutine):
        super().__init__()
        assert isinstance(
            context,
            interactive_widgets.backend.contexts.docker_context.DockerContext,
        )

        self.context = context
        self.configuration = configuration
        self.name = name
        self.send_message = send_message

        self.logger = logging.getLogger(self.configuration['logger_name'])
        self.volume: typing.Optional[aiodocker.docker.DockerVolume] = None

    async def instantiate(self, volume: aiodocker.docker.DockerVolume):
        self.volume = volume

    async def handle_message(self, message: typing.Any):
        raise NotImplementedError

    async def tear_down(self):
        pass

    async def _run_once(self):
        try:
            self.logger.debug('Creating container...')
            try:
                container = await interactive_widgets.backend.shield.shield(
                    self.context.docker.containers.create(
                        config={
                            'Cmd': self.configuration['command'],
                            'Image': self.configuration['image'],
                            'HostConfig': {
                                'Mounts': [
                                    {
                                        'Target': '/data',
                                        'Source': self.volume.name,
                                        'Type': 'volume',
                                        # TODO: 'VolumeOptions' labels
                                    },
                                ],
                            },
                        },
                        name=f'interactive_widgets_{binascii.hexlify(self.name.encode("utf-8")).decode("utf-8")}',
                    ),
                )
            except:
                self.logger.debug('Reverting container creation...')
                container = aiodocker.containers.DockerContainer(
                    self.context.docker,
                    id=f'interactive_widgets_{binascii.hexlify(self.name.encode("utf-8")).decode("utf-8")}',
                )
                try:
                    await container.delete(force=True)
                except aiodocker.exceptions.DockerError as error:
                    if error.status != 404:
                        raise
                    self.logger.debug('Container had not been created yet.')
                self.container = None
                raise

            self.logger.debug('Attaching to container...')
            async with container.attach(stdout=True, stderr=True, logs=True) as stream:
                self.logger.debug('Starting container...')
                await container.start()
                while True:
                    message = await stream.read_out()
                    if message is None:
                        break
                    assert message.stream in [1, 2]
                    await self.send_message({
                        'type': 'output',
                        'stdout' if message.stream == 1 else 'stderr': base64.b64encode(message.data).decode('utf-8')
                    })
        finally:
            self.logger.debug('Stopping container...')
            await container.stop()
            self.logger.debug('Deleting container...')
            await container.delete(force=True)

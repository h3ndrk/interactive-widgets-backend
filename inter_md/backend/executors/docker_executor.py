import abc
import aiodocker
import base64
import binascii
import collections
import logging
import typing

from .. import contexts


class DockerExecutor(abc.ABC):

    def __init__(self, context: contexts.Context, configuration: dict, name: str, send_message: collections.abc.Coroutine):
        super().__init__()
        assert isinstance(
            context,
            contexts.DockerContext,
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
            container = await self.context.docker.containers.create(
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
                name=f'inter_md_{binascii.hexlify(self.name.encode("utf-8")).decode("utf-8")}'
            )

            self.logger.debug('Starting container...')
            await container.start()

            self.logger.debug('Attaching to container...')
            async with container.attach(stdout=True, stderr=True, logs=True) as stream:
                while True:
                    message = await stream.read_out()
                    if message is None:
                        break
                    assert message.stream in [1, 2]

                    await self.send_message({
                        'stdout' if message.stream == 1 else 'stderr': base64.b64encode(message.data).decode('utf-8')
                    })
        finally:
            self.logger.debug('Deleting container...')
            await container.delete(force=True)

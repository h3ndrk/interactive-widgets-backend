import aiodocker
import asyncio
import binascii
import collections
import typing

from .docker_executor import DockerExecutor


class Initialization(DockerExecutor):

    def __init__(self, docker: aiodocker.Docker, name: str, image_name: str, command: typing.Iterable[str], send_message: collections.abc.Coroutine):
        super().__init__()
        self.docker = docker
        self.name = name
        self.image_name = image_name
        self.command = command
        self.send_message = send_message

    async def instantiate(self, *args, **kwargs):
        await super().instantiate(*args, **kwargs)

        try:
            container = await self.docker.containers.create(
                config={
                    'Cmd': self.command,
                    'Image': self.image_name,
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
            await container.start()
            async with container.attach(stdout=True, stderr=True, logs=True) as stream:
                while True:
                    message = await stream.read_out()
                    if message is None:
                        break
                    print(f'received: {message}')
        finally:
            await container.delete(force=True)

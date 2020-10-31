import aiodocker
import asyncio
import base64
import binascii
import collections
import logging
import traceback
import typing

from .. import executors


async def docker_once(docker: aiodocker.Docker, name: str, command: typing.List[str], image: str, volume_name: str, logger: logging.Logger, write_output: collections.abc.Coroutine):
    try:
        logger.debug('Creating container...')
        container = await docker.containers.create(
            config={
                'Cmd': command,
                'Image': image,
                'HostConfig': {
                    'Mounts': [
                        {
                            'Target': '/data',
                            'Source': volume_name,
                            'Type': 'volume',
                            # TODO: 'VolumeOptions' labels
                        },
                    ],
                },
            },
            name=f'inter_md_{binascii.hexlify(name.encode("utf-8")).decode("utf-8")}'
        )

        logger.debug('Starting container...')
        await container.start()

        logger.debug('Attaching to container...')
        async with container.attach(stdout=True, stderr=True, logs=True) as stream:
            while True:
                message = await stream.read_out()
                if message is None:
                    break
                assert message.stream in [1, 2]

                await write_output({
                    'stdout' if message.stream == 1 else 'stderr': base64.b64encode(message.data).decode('utf-8')
                })
    finally:
        logger.debug('Deleting container...')
        await container.delete(force=True)


class DockerOnce(executors.DockerExecutor):

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

        self.execute_task: typing.Optional[asyncio.Task] = None

    async def _execute(self):
        try:
            await self.send_message({
                'triggered': True,
            })
            await executors.docker_once(
                self.context.docker,
                self.name,
                self.configuration['command'],
                self.configuration['image'],
                self.volume.name,
                self.logger,
                self.send_message,
            )
        except:
            await self.send_message({
                'stderr': base64.b64encode(traceback.format_exc().encode('utf-8')).decode('utf-8'),
            })
        finally:
            self.execute_task = None

    async def handle_message(self, message: typing.Any):
        self.execute_task = asyncio.create_task(self._execute())

    async def tear_down(self):
        if self.execute_task is not None:
            self.execute_task.cancel()
            try:
                await self.execute_task
            except asyncio.CancelledError:
                pass

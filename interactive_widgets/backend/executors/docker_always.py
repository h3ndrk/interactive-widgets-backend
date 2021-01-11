import aiodocker
import asyncio
import base64
import binascii
import typing

import interactive_widgets.backend.executors.docker_executor


class DockerAlways(interactive_widgets.backend.executors.docker_executor.DockerExecutor):

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

        self.execute_task: typing.Optional[asyncio.Task] = None
        self.container: typing.Optional[aiodocker.docker.DockerContainer] = None
        self.stream: typing.Optional[aiodocker.stream.Stream] = None
        self.stream_ready = asyncio.Event()

    async def instantiate(self, *args, **kwargs):
        await super().instantiate(*args, **kwargs)

        self.execute_task = asyncio.create_task(self._execute())

    async def _execute(self):
        while True:
            try:
                self.logger.debug('Creating container...')
                self.container = await self.context.docker.containers.create(
                    config={
                        'Cmd': self.configuration['command'],
                        'Image': self.configuration['image'],
                        'AttachStdin': True,
                        'Tty': self.configuration.get('enable_tty', False),
                        'OpenStdin': True,
                        'StdinOnce': True,
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
                    name=f'interactive_widgets_{binascii.hexlify(self.name.encode("utf-8")).decode("utf-8")}'
                )

                self.logger.debug('Attaching to container...')
                async with self.container.attach(stdin=True, stdout=True, stderr=True, logs=True) as stream:
                    self.stream = stream
                    try:
                        self.logger.debug('Starting container...')
                        await self.container.start()
                        self.stream_ready.set()
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
                        self.stream_ready.clear()
                        self.stream = None
            finally:
                self.logger.debug('Stopping container...')
                await self.container.stop()
                self.logger.debug('Deleting container...')
                await self.container.delete(force=True)
                self.container = None

    async def handle_message(self, message: dict):
        await asyncio.wait_for(self.stream_ready.wait(), self.configuration.get('handle_message_timeout', 10))
        assert self.container is not None
        if 'stdin' in message:
            await self.stream.write_in(
                base64.b64decode(message['stdin'].encode('utf-8')),
            )
        elif 'size' in message:
            print(message['size'])
            await self.container.resize(
                h=message['size']['rows'],
                w=message['size']['cols'],
            )
        else:
            raise NotImplementedError

    async def tear_down(self):
        if self.execute_task is not None:
            self.execute_task.cancel()
            self.logger.debug('Waiting for execute task...')
            try:
                await self.execute_task
            except asyncio.CancelledError:
                pass

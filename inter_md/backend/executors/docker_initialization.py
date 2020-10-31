import aiodocker
import asyncio
import base64
import binascii
import collections
import typing

import inter_md.backend.executors


class DockerInitialization(inter_md.backend.executors.DockerExecutor):

    async def instantiate(self, *args, **kwargs):
        await super().instantiate(*args, **kwargs)

        try:
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

            await container.start()

            async with container.attach(stdout=True, stderr=True, logs=True) as stream:
                while True:
                    message = await stream.read_out()
                    if message is None:
                        break
                    assert message.stream in [1, 2]

                    await self.send_message(
                        self.name,
                        {
                            'stdout' if message.stream == 1 else 'stderr': base64.b64encode(message.data).decode('utf-8')
                        },
                    )
        finally:
            await container.delete(force=True)

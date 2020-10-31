import abc
import aiodocker
import collections
import typing

import inter_md.backend.contexts


class DockerExecutor(abc.ABC):

    def __init__(self, context: inter_md.backend.contexts.Context, configuration: dict, name: str, send_message: collections.abc.Coroutine):
        super().__init__()
        assert isinstance(
            context,
            inter_md.backend.contexts.DockerContext,
        )

        self.context = context
        self.configuration = configuration
        self.name = name
        self.send_message = send_message
        self.volume: typing.Optional[aiodocker.docker.DockerVolume] = None

    async def instantiate(self, volume: aiodocker.docker.DockerVolume):
        self.volume = volume

    async def handle_message(self, message: typing.Any):
        raise NotImplementedError

    async def tear_down(self):
        pass

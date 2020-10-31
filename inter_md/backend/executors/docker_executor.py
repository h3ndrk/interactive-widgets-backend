import abc
import aiodocker
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

import abc
import aiodocker
import typing


class DockerExecutor(abc.ABC):

    def __init__(self):
        super().__init__()
        self.volume: typing.Optional[aiodocker.docker.DockerVolume] = None

    async def instantiate(self, volume: aiodocker.docker.DockerVolume):
        self.volume = volume

    async def on_message(self, message: typing.Any):
        pass

    async def tear_down(self):
        pass

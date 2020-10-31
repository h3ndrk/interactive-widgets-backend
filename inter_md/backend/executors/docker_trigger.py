import typing

from .. import executors


class DockerTrigger(executors.DockerExecutor):

    async def handle_message(self, message: typing.Any):
        await executors.docker_once(
            self.context.docker,
            self.name,
            self.configuration['command'],
            self.configuration['image'],
            self.volume.name,
            self.logger,
            self.send_message,
        )

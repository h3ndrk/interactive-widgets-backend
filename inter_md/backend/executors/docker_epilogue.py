from .. import executors


class DockerEpilogue(executors.DockerExecutor):

    async def tear_down(self):
        await executors.docker_once(
            self.context.docker,
            self.name,
            self.configuration['command'],
            self.configuration['image'],
            self.volume.name,
            self.logger,
            self.send_message,
        )

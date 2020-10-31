from .. import executors


class DockerPrologue(executors.DockerExecutor):

    async def instantiate(self, *args, **kwargs):
        await super().instantiate(*args, **kwargs)
        
        await executors.docker_once(
            self.context.docker,
            self.name,
            self.configuration['command'],
            self.configuration['image'],
            self.volume.name,
            self.logger,
            self.send_message,
        )

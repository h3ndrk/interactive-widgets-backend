from .. import executors


class DockerPrologue(executors.DockerExecutor):

    async def instantiate(self, *args, **kwargs):
        await super().instantiate(*args, **kwargs)

        await self.run_once()

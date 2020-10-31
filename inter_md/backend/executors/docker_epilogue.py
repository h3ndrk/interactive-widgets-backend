from .. import executors


class DockerEpilogue(executors.DockerExecutor):

    async def tear_down(self):
        await self.run_once()

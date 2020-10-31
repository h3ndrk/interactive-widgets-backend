from .. import executors


class DockerEpilogue(executors.DockerExecutor):

    async def tear_down(self):
        await self._run_once()

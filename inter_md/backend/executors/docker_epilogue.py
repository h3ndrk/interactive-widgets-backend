import inter_md.backend.executors.docker_executor


class DockerEpilogue(inter_md.backend.executors.docker_executor.DockerExecutor):

    async def tear_down(self):
        await self._run_once()

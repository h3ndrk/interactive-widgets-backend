import inter_md.backend.executors.docker_executor


class DockerPrologue(inter_md.backend.executors.docker_executor.DockerExecutor):

    async def instantiate(self, *args, **kwargs):
        await super().instantiate(*args, **kwargs)

        await self._run_once()

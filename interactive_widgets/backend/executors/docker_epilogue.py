import interactive_widgets.backend.executors.docker_executor


class DockerEpilogue(interactive_widgets.backend.executors.docker_executor.DockerExecutor):

    async def tear_down(self):
        await self._run_once()

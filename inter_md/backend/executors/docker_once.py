import asyncio
import base64
import traceback
import typing

import inter_md.backend.executors.docker_executor


class DockerOnce(inter_md.backend.executors.docker_executor.DockerExecutor):

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

        self.execute_task: typing.Optional[asyncio.Task] = None

    async def _execute(self):
        try:
            await self.send_message({
                'type': 'started',
            })
            await self._run_once()
            await self.send_message({
                'type': 'finished',
            })
        except:
            await self.send_message({
                'type': 'errored',
                'message': base64.b64encode(traceback.format_exc().encode('utf-8')).decode('utf-8'),
            })
        finally:
            self.execute_task = None

    async def handle_message(self, message: typing.Any):
        # TODO: enhance what happens if a task already executes
        if self.execute_task is None:
            self.execute_task = asyncio.create_task(self._execute())

    async def tear_down(self):
        if self.execute_task is not None:
            self.execute_task.cancel()
            try:
                await self.execute_task
            except asyncio.CancelledError:
                pass

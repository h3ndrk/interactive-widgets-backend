import asyncio
import base64
import traceback
import typing

from .. import executors


class DockerTrigger(executors.DockerExecutor):

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

        self.execute_task: typing.Optional[asyncio.Task] = None

    async def _execute(self):
        try:
            await executors.docker_once(
                self.context.docker,
                self.name,
                self.configuration['command'],
                self.configuration['image'],
                self.volume.name,
                self.logger,
                self.send_message,
            )
        except:
            await self.send_message({
                'stderr': base64.b64encode(traceback.format_exc().encode('utf-8')).decode('utf-8'),
            })
        finally:
            self.execute_task = None

    async def handle_message(self, message: typing.Any):
        self.execute_task = asyncio.create_task(self._execute())

    async def tear_down(self):
        if self.execute_task is not None:
            self.execute_task.cancel()
            try:
                await self.execute_task
            except asyncio.CancelledError:
                pass

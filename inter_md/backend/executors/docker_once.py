import aiodocker
import asyncio
import base64
import binascii
import collections
import logging
import traceback
import typing

from .. import executors





class DockerOnce(executors.DockerExecutor):

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

        self.execute_task: typing.Optional[asyncio.Task] = None

    async def _execute(self):
        try:
            await self.send_message({
                'triggered': True,
            })
            await self._run_once()
        except:
            await self.send_message({
                'stderr': base64.b64encode(traceback.format_exc().encode('utf-8')).decode('utf-8'),
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

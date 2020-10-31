import aiodocker
import aiohttp.web
import asyncio
import typing

from .page import Page


class Server:

    def __init__(self, docker: aiodocker.Docker):
        self.docker = docker
        self.application = aiohttp.web.Application()
        self.pages: typing.Dict[str, Page] = {
            '/test': Page(self.docker),
        }
        for url, page in self.pages.items():
            self.application.add_subapp(url, page.application)

    async def run(self, host: str, port: int):
        runner = aiohttp.web.AppRunner(
            self.application,
            handle_signals=True,
            access_log=None,
        )
        await runner.setup()

        site = aiohttp.web.TCPSite(
            runner=runner,
            host=host,
            port=port,
        )
        await site.start()

        eternity_event = asyncio.Event()
        try:
            for site in runner.sites:
                print(f'Listening on {str(site.name)}...')
            await eternity_event.wait()
        finally:
            await runner.cleanup()

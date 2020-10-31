import aiodocker
import aiohttp.web
import asyncio
import typing

import inter_md.backend
import inter_md.backend.contexts


class Server:

    def __init__(self, configuration: dict):
        self.configuration = configuration

    async def run(self):
        async with getattr(inter_md.backend.contexts, self.configuration['context']['type'])(self.configuration['context']) as context:
            application = aiohttp.web.Application()

            pages = {}
            for url, page_configuration in self.configuration['pages'].items():
                pages[url] = inter_md.backend.Page(context, page_configuration)
                application.add_subapp(url, pages[url].application)

            runner = aiohttp.web.AppRunner(
                application,
                handle_signals=True,
                access_log=None,
            )
            await runner.setup()

            site = aiohttp.web.TCPSite(
                runner=runner,
                host=self.configuration['host'],
                port=self.configuration['port'],
            )
            await site.start()

            eternity_event = asyncio.Event()
            try:
                for site in runner.sites:
                    print(f'Listening on {str(site.name)}...')
                await eternity_event.wait()
            finally:
                await runner.cleanup()

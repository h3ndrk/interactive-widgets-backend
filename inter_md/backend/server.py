import aiodocker
import aiohttp.web
import asyncio
import logging
import typing

from .. import backend
from . import contexts


class Server:

    def __init__(self, configuration: dict):
        self.configuration = configuration
        self.logger = logging.getLogger(self.configuration['logger_name'])

    async def run(self):
        async with getattr(contexts, self.configuration['context']['type'])(self.configuration['context']) as context:
            application = aiohttp.web.Application()

            self.logger.debug('Adding pages...')
            pages = {}
            for url, page_configuration in self.configuration['pages'].items():
                self.logger.debug(f'Adding page {url}...')
                pages[url] = backend.Page(context, page_configuration)
                application.add_subapp(url, pages[url].application)
            self.logger.debug('Pages added.')

            self.logger.debug('Starting server...')
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
                    self.logger.info(f'Listening on {str(site.name)}...')
                await eternity_event.wait()
            finally:
                await runner.cleanup()

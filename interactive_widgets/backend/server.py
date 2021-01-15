import aiodocker
import aiohttp.web
import asyncio
import logging
import pathlib
import signal
import typing

import interactive_widgets.backend.page
import interactive_widgets.backend.contexts.get


class Server:

    def __init__(self, configuration: dict):
        self.configuration = configuration
        self.logger = logging.getLogger(self.configuration['logger_name'])
        self.eternity_event = asyncio.Event()

    def stop(self):
        self.eternity_event.set()

    async def run(self):
        loop = asyncio.get_running_loop()
        loop.add_signal_handler(signal.SIGINT, self.stop)
        loop.add_signal_handler(signal.SIGTERM, self.stop)

        async with interactive_widgets.backend.contexts.get.get(self.configuration['context']['type'])(self.configuration['context']) as context:
            application = aiohttp.web.Application()

            self.logger.debug('Adding pages...')
            pages = {}
            for url, page_configuration in self.configuration['pages'].items():
                self.logger.debug(f'Adding page {url}...')
                pages[url] = interactive_widgets.backend.page.Page(
                    context,
                    page_configuration,
                    pathlib.PurePosixPath(url),
                    application,
                )
            self.logger.debug('Pages added.')

            self.logger.debug('Starting server...')
            runner = aiohttp.web.AppRunner(
                application,
                access_log=None,
            )
            await runner.setup()

            site = aiohttp.web.TCPSite(
                runner=runner,
                host=self.configuration['host'],
                port=self.configuration['port'],
                shutdown_timeout=1,
            )
            await site.start()

            try:
                for site in runner.sites:
                    self.logger.info(f'Listening on {str(site.name)}...')
                await self.eternity_event.wait()
            except asyncio.CancelledError:
                pass
            finally:
                await runner.cleanup()

import asyncio
import click
import json
import logging

import interactive_widgets.backend.server


async def async_main(**arguments):
    configuration = json.load(arguments['configuration'])

    logging.basicConfig(
        level=configuration['logging_level'],
        format='%(asctime)s  %(name)-20s  %(levelname)-8s  %(message)s',
    )

    server = interactive_widgets.backend.server.Server(configuration)
    await server.run()


@click.command()
@click.argument('configuration', type=click.File())
def main(**arguments):
    asyncio.run(async_main(**arguments))

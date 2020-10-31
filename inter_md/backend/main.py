import asyncio
import click
import json
import logging

from .. import backend


logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s  %(name)-17s  %(levelname)-8s  %(message)s',
)


async def async_main(**arguments):
    server = backend.Server(json.load(arguments['configuration']))
    await server.run()


@click.command()
@click.argument('configuration', type=click.File())
def main(**arguments):
    asyncio.run(async_main(**arguments))

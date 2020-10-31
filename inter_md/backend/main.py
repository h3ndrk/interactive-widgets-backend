import click
import aiodocker
import aiohttp.web
import asyncio
import logging
import typing

from .room import Room
from .docker_room import DockerRoom
from .page import Page
from .server import Server

logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s  %(name)-17s  %(levelname)-8s  %(message)s',
)


async def async_main(**arguments):
    async with aiodocker.Docker() as docker:
        server = Server(docker)
        await server.run(arguments['host'], arguments['port'])


@click.command()
@click.option('--host', default='*', help='Hostname to listen on', show_default=True)
@click.option('--port', default=8080, help='Port to listen on', show_default=True)
def main(**arguments):
    asyncio.run(async_main(**arguments))

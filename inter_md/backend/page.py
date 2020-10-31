import aiodocker
import typing

from .room import Room
from .room_connection import RoomConnection


class Page:

    def __init__(self, docker: aiodocker.Docker):
        self.docker = docker
        self.rooms: typing.Dict[str, Room] = {}

    def connect(self, room_name: str, websocket):
        return RoomConnection(self.docker, self.rooms, room_name, websocket)

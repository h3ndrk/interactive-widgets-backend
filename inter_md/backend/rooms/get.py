import inter_md.backend.rooms.docker_room


def get(name: str):
    return {
        'docker': inter_md.backend.rooms.docker_room.DockerRoom,
    }[name]

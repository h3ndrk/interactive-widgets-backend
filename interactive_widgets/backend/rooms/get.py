import interactive_widgets.backend.rooms.docker_room


def get(name: str):
    return {
        'docker': interactive_widgets.backend.rooms.docker_room.DockerRoom,
    }[name]

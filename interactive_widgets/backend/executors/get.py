import interactive_widgets.backend.executors.docker_always
import interactive_widgets.backend.executors.docker_epilogue
import interactive_widgets.backend.executors.docker_once
import interactive_widgets.backend.executors.docker_prologue


def get(name: str):
    return {
        'docker.always': interactive_widgets.backend.executors.docker_always.DockerAlways,
        'docker.epilogue': interactive_widgets.backend.executors.docker_epilogue.DockerEpilogue,
        'docker.once': interactive_widgets.backend.executors.docker_once.DockerOnce,
        'docker.prologue': interactive_widgets.backend.executors.docker_prologue.DockerPrologue,
    }[name]

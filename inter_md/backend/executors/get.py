import inter_md.backend.executors.docker_always
import inter_md.backend.executors.docker_epilogue
import inter_md.backend.executors.docker_once
import inter_md.backend.executors.docker_prologue


def get(name: str):
    return {
        'docker.always': inter_md.backend.executors.docker_always.DockerAlways,
        'docker.epilogue': inter_md.backend.executors.docker_epilogue.DockerEpilogue,
        'docker.once': inter_md.backend.executors.docker_once.DockerOnce,
        'docker.prologue': inter_md.backend.executors.docker_prologue.DockerPrologue,
    }[name]

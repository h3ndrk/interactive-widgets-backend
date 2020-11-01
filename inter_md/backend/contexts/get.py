import inter_md.backend.contexts.docker_context


def get(name: str):
    return {
        'docker': inter_md.backend.contexts.docker_context.DockerContext,
    }[name]

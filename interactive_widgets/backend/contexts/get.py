import interactive_widgets.backend.contexts.docker_context


def get(name: str):
    return {
        'docker': interactive_widgets.backend.contexts.docker_context.DockerContext,
    }[name]

import aiodocker

import interactive_widgets.backend.contexts.context


class DockerContext(interactive_widgets.backend.contexts.context.Context):

    async def __aenter__(self):
        self.docker = aiodocker.Docker(
            url=self.configuration.get('url', None),
        )
        await self.docker.__aenter__()
        self.logger.debug(await self.docker.version())
        return self

    async def __aexit__(self, *args, **kwargs):
        await self.docker.__aexit__(*args, **kwargs)

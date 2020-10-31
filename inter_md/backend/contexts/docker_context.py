import aiodocker

import inter_md.backend.contexts


class DockerContext(inter_md.backend.contexts.Context):

    async def __aenter__(self):
        self.docker = aiodocker.Docker(
            url=self.configuration['url'],
        )
        await self.docker.__aenter__()
        return self

    async def __aexit__(self, *args, **kwargs):
        await self.docker.__aexit__(*args, **kwargs)

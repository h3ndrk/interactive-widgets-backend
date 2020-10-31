import abc


class Context(abc.ABC):

    def __init__(self, configuration: dict):
        self.configuration = configuration

    @abc.abstractmethod
    async def __aenter__(self):
        raise NotImplementedError

    @abc.abstractmethod
    async def __aexit__(self, *args, **kwargs):
        raise NotImplementedError

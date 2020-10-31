import abc
import logging


class Context(abc.ABC):

    def __init__(self, configuration: dict):
        self.configuration = configuration
        self.logger = logging.getLogger(self.configuration['logger_name'])

    @abc.abstractmethod
    async def __aenter__(self):
        raise NotImplementedError

    @abc.abstractmethod
    async def __aexit__(self, *args, **kwargs):
        raise NotImplementedError

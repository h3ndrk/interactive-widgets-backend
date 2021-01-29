# interactive-widgets-backend

This repository contains the backend powering *interactive-widgets* projects ([MkDocs plugin](https://github.com/h3ndrk/interactive-widgets-mkdocs/)). It reads a configuration and starts a HTTP server where it reacts to incoming WebSocket connections to instantiate rooms based on Docker containers.

## Concept

**Pages** are identified by URL and may be instantiated into **rooms** which are identified by name. Multiple clients can connect to the same room (when connecting with the same room name), or connect to separate, isolated rooms (when connecting with different room names). The first incoming client for a non-existing room makes the backend creating and instantiating it. The last disconnecting client of a room makes the backend tearing it down and removing it.

Pages consist of **executors** which correspond to the execution part of *widgets* in *interactive-widgets-mkdocs*. When a room is instantiated from a page, also the executors are instantiated. The same holds true for the tear down.

Consider one room with many connected clients: Data being sent from a client to the backend is not echoed to other connected clients. Data being sent from the backend to connected clients is broadcasted to all connected clients. (Data is only sent within the communication members of one room, not across rooms.)

## Docker Implementation

Currently, the backend only implements rooms and executors based on Docker. A room corresponds to a shared Docker volume, executors within a room correspond to Docker containers attached to the shared room volume.

There exist several executor types: *always* (long-running, restarting container), *epilogue* (short-lived container executed at tear down), *once* (short-lived container startable via client connection), *prologue* (short-lived container executed at instantiation)

The backend does not build nor pull any Docker images. All images used by executors must be present beforehand.

## Building

For development you may install the backend and monitor executable via `pip`, e.g.: `pip install --editable ./`, then run the executables `interactive-widgets-backend` and `interactive-widgets-monitor`.

For usage with the [MkDocs plugin](https://github.com/h3ndrk/interactive-widgets-mkdocs/) the executables can be built with: `docker-compose build` (creates Docker images `interactive-widgets-backend` and `interactive-widgets-monitor`).

## License

MIT

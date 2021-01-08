import base64
import pathlib


def write(path: pathlib.Path, message: dict):
    if 'contents' in message:
        path.parent.mkdir(parents=True, exist_ok=True)
        with path.open(mode='wb') as file:
            file.write(base64.b64decode(message['contents']))
    elif 'delete' in message:
        path.unlink()

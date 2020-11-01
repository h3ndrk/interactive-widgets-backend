import base64
import pathlib
import sys


def read(path: pathlib.Path):
    with path.open(mode='rb') as file:
        sys.stdout.write(
            f'{{"contents":"{base64.b64encode(file.read()).decode("utf-8")}"}}\n',
        )

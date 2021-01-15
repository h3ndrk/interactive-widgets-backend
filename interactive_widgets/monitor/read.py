import base64
import pathlib
import sys
import threading


def read(path: pathlib.Path, stdout_lock: threading.Lock):
    with path.open(mode='rb') as file, stdout_lock:
        sys.stdout.write(
            f'{{"contents":"{base64.b64encode(file.read()).decode("utf-8")}"}}\n',
        )
        sys.stdout.flush()

import base64
import json
import pathlib
import signal
import sys
import threading
import time
import traceback

import interactive_widgets.monitor.read
import interactive_widgets.monitor.wait_for_event
import interactive_widgets.monitor.write


class StdinReader(threading.Thread):

    def __init__(self, path: pathlib.Path, stdout_lock: threading.Lock, *args, **kwargs):
        super().__init__(*args, **kwargs)

        self.daemon = True
        self.path = path
        self.stdout_lock = stdout_lock

    def run(self):
        for line in sys.stdin:
            try:
                interactive_widgets.monitor.write.write(
                    self.path,
                    json.loads(line),
                )
            except SystemExit:
                return
            except OSError as error:
                with self.stdout_lock:
                    sys.stdout.write(
                        f'{{"error":"{base64.b64encode(error.strerror.encode("utf-8")).decode("utf-8")}"}}\n',
                    )
            except:
                with self.stdout_lock:
                    sys.stdout.write(
                        f'{{"error":"{base64.b64encode(traceback.format_exc().encode("utf-8")).decode("utf-8")}"}}\n',
                    )


def main():
    if len(sys.argv) != 4:
        sys.stderr.write(
            f'Usage: {sys.argv[0]} FILE SUCCESS_TIMEOUT FAILURE_TIMEOUT\n')
        sys.exit(1)

    signal.signal(signal.SIGINT, lambda *unused: sys.exit())
    signal.signal(signal.SIGTERM, lambda *unused: sys.exit())

    file = sys.argv[1]
    success_timeout = float(sys.argv[2])
    failure_timeout = float(sys.argv[3])

    stdout_lock = threading.Lock()

    StdinReader(pathlib.Path(file), stdout_lock).start()

    while True:
        try:
            interactive_widgets.monitor.read.read(
                pathlib.Path(file),
                stdout_lock,
            )
        except SystemExit:
            return
        except OSError as error:
            with stdout_lock:
                sys.stdout.write(
                    f'{{"error":"{base64.b64encode(error.strerror.encode("utf-8")).decode("utf-8")}"}}\n',
                )
            time.sleep(failure_timeout)
            continue
        except:
            with stdout_lock:
                sys.stdout.write(
                    f'{{"error":"{base64.b64encode(traceback.format_exc().encode("utf-8")).decode("utf-8")}"}}\n',
                )
            time.sleep(failure_timeout)
            continue
        try:
            interactive_widgets.monitor.wait_for_event.wait_for_event(
                pathlib.Path(file),
            )
            time.sleep(success_timeout)
        except SystemExit:
            return
        except:
            with stdout_lock:
                sys.stdout.write(
                    f'{{"error":"{base64.b64encode(traceback.format_exc().encode("utf-8")).decode("utf-8")}"}}\n',
                )
            time.sleep(failure_timeout)

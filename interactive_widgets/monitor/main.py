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

    def __init__(self, path: pathlib.Path, stdout_lock: threading.Lock, written_event: threading.Event, *args, **kwargs):
        super().__init__(*args, **kwargs)

        self.daemon = True
        self.path = path
        self.stdout_lock = stdout_lock
        self.written_event = written_event

    def run(self):
        for line in sys.stdin:
            try:
                interactive_widgets.monitor.write.write(
                    self.path,
                    json.loads(line),
                )
                self.written_event.set()
                self.written_event.clear()
            except SystemExit:
                return
            except OSError as error:
                with self.stdout_lock:
                    sys.stdout.write(
                        f'{{"error":"{base64.b64encode(error.strerror.encode("utf-8")).decode("utf-8")}"}}\n',
                    )
                    sys.stdout.flush()
            except:
                with self.stdout_lock:
                    sys.stdout.write(
                        f'{{"error":"{base64.b64encode(traceback.format_exc().encode("utf-8")).decode("utf-8")}"}}\n',
                    )
                    sys.stdout.flush()


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
    written_event = threading.Event()

    StdinReader(pathlib.Path(file), stdout_lock, written_event).start()

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
                sys.stdout.flush()
            written_event.wait(failure_timeout)
            continue
        except:
            with stdout_lock:
                sys.stdout.write(
                    f'{{"error":"{base64.b64encode(traceback.format_exc().encode("utf-8")).decode("utf-8")}"}}\n',
                )
                sys.stdout.flush()
            written_event.wait(failure_timeout)
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
                sys.stdout.flush()
            written_event.wait(failure_timeout)

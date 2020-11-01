import base64
import click
import json
import pathlib
import sys
import threading
import time
import traceback

from .. import monitor


class StdinReader(threading.Thread):

    def __init__(self, path: pathlib.Path, *args, **kwargs):
        super().__init__(*args, **kwargs)

        self.daemon = True
        self.path = path

    def run(self):
        for line in sys.stdin:
            try:
                monitor.write(self.path, json.loads(line))
            except KeyboardInterrupt:
                return
            except OSError as error:
                sys.stdout.write(
                    f'{{"error":"{base64.b64encode(error.strerror.encode("utf-8")).decode("utf-8")}"}}\n',
                )
            except:
                sys.stdout.write(
                    f'{{"error":"{base64.b64encode(traceback.format_exc().encode("utf-8")).decode("utf-8")}"}}\n',
                )


@click.command()
@click.option('--success-timeout', default=0.1, help='Timeout between monitoring event and read operation in seconds', show_default=True)
@click.option('--failure-timeout', default=5.0, help='Timeout between a failure and next read operation in seconds', show_default=True)
@click.argument('file', type=click.Path())
def main(**arguments):
    StdinReader(pathlib.Path(arguments['file'])).start()

    while True:
        try:
            monitor.read(pathlib.Path(arguments['file']))
        except KeyboardInterrupt:
            return
        except OSError as error:
            sys.stdout.write(
                f'{{"error":"{base64.b64encode(error.strerror.encode("utf-8")).decode("utf-8")}"}}\n',
            )
            time.sleep(arguments['failure_timeout'])
            continue
        except:
            sys.stdout.write(
                f'{{"error":"{base64.b64encode(traceback.format_exc().encode("utf-8")).decode("utf-8")}"}}\n',
            )
            time.sleep(arguments['failure_timeout'])
            continue
        try:
            monitor.wait_for_event(pathlib.Path(arguments['file']))
            time.sleep(arguments['success_timeout'])
        except KeyboardInterrupt:
            return
        except:
            sys.stdout.write(
                f'{{"error":"{base64.b64encode(traceback.format_exc().encode("utf-8")).decode("utf-8")}"}}\n',
            )
            time.sleep(arguments['failure_timeout'])

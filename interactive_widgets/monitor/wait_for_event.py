import inotify.adapters
import pathlib


def parent_paths(path: pathlib.Path):
    current_path = path
    while current_path.parent != current_path:
        current_path = current_path.parent
        yield current_path


def wait_for_event(path: pathlib.Path):
    paths = [path] + list(parent_paths(path))
    adapter = inotify.adapters.Inotify(
        paths=[str(path) for path in paths],
        block_duration_s=60,
    )

    matching_type_names = set([
        'IN_CLOSE_WRITE',
        'IN_DELETE',
        'IN_DELETE_SELF',
        'IN_MODIFY',
        'IN_MOVED_FROM',
        'IN_MOVE_SELF',
    ])
    while True:
        for event in adapter.event_gen(yield_nones=False):
            (_, event_type_names, event_path, event_filename) = event
            if len(set(event_type_names) & matching_type_names) == 0:
                continue

            path = pathlib.Path(event_path)
            if len(event_filename) > 0:
                path /= event_filename
            if path in paths:
                return

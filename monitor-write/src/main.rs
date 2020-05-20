extern crate base64;
extern crate inotify;
extern crate signal_hook;

fn parent_paths_from_child(child_path: &std::path::PathBuf) -> Vec<std::path::PathBuf> {
    let mut parent_paths: Vec<std::path::PathBuf> = vec![];
    let mut current_path = child_path.as_path();
    while current_path.parent().is_some() {
        current_path = current_path.parent().unwrap();
        parent_paths.push(current_path.to_path_buf());
    }
    parent_paths
}

fn add_watches(
    i: &mut inotify::Inotify,
    child_path: &std::path::PathBuf,
    parent_paths: &[std::path::PathBuf],
) -> std::io::Result<std::collections::HashMap<inotify::WatchDescriptor, std::path::PathBuf>> {
    let mut watch_descriptors: std::collections::HashMap<
        inotify::WatchDescriptor,
        std::path::PathBuf,
    > = std::collections::HashMap::new();
    watch_descriptors.insert(
        i.add_watch(
            child_path,
            inotify::WatchMask::CREATE
                | inotify::WatchMask::CLOSE_WRITE
                | inotify::WatchMask::MODIFY
                | inotify::WatchMask::MOVE_SELF
                | inotify::WatchMask::DELETE_SELF,
        )?,
        child_path.to_path_buf(),
    );
    for parent_path in parent_paths {
        watch_descriptors.insert(
            i.add_watch(parent_path, inotify::WatchMask::MOVED_FROM)?,
            parent_path.to_path_buf(),
        );
    }
    Ok(watch_descriptors)
}

fn rm_watches(
    i: &mut inotify::Inotify,
    watch_descriptors: std::collections::HashMap<inotify::WatchDescriptor, std::path::PathBuf>,
) -> std::io::Result<()> {
    for (watch_descriptor, _) in watch_descriptors {
        i.rm_watch(watch_descriptor)?;
    }
    Ok(())
}

fn wait_for_event(
    child_path: &std::path::PathBuf,
    parent_paths: &[std::path::PathBuf],
) -> std::io::Result<()> {
    let mut i = inotify::Inotify::init()?;
    let watch_descriptors = add_watches(&mut i, child_path, parent_paths)?;
    let mut buffer = [0u8; 1024];
    loop {
        let events = match i.read_events_blocking(&mut buffer) {
            Ok(events) => events,
            Err(e) => {
                let _ = rm_watches(&mut i, watch_descriptors);
                return Err(e);
            }
        };
        let mut at_least_one_path_changed = false;
        for event in events {
            let path_from_watch_descriptor = match watch_descriptors.get(&event.wd) {
                Some(watch_descriptor) => watch_descriptor,
                None => {
                    return Err(std::io::Error::new(
                        std::io::ErrorKind::Other,
                        "unexpected watch descriptor",
                    ))
                }
            };
            // `event.name` is only Some if watch is directory -> append changed path
            let changed_path = match event.name {
                Some(changed_path) => path_from_watch_descriptor.join(changed_path),
                None => path_from_watch_descriptor.to_path_buf(),
            };
            if child_path == &changed_path {
                at_least_one_path_changed = true;
                break;
            }
            if parent_paths
                .iter()
                .any(|parent_path| parent_path.as_path() == changed_path)
            {
                at_least_one_path_changed = true;
                break;
            }
        }
        if at_least_one_path_changed {
            break;
        }
    }
    rm_watches(&mut i, watch_descriptors)?;
    Ok(())
}

fn read_file(child_path: &std::path::PathBuf) {
    let mut last_contents = String::new();
    let parent_paths = parent_paths_from_child(child_path);
    loop {
        let mut got_error = false;
        let contents = match std::fs::read(child_path) {
            Ok(bytes) => base64::encode(bytes),
            Err(error) => {
                match error.kind() {
                    std::io::ErrorKind::NotFound => {}
                    _ => {
                        eprintln!("Failed to read file: {}", error);
                    }
                }
                got_error = true;
                String::new()
            }
        };
        if contents != last_contents {
            println!("{}", contents);
            last_contents = contents;
        }
        if !got_error {
            if let Err(error) = wait_for_event(child_path, &parent_paths) {
                eprintln!("Failed to wait for events: {}", error);
                got_error = true;
            }
        }
        if got_error {
            std::thread::sleep(std::time::Duration::from_secs(5));
        }
    }
}

fn main() -> std::io::Result<()> {
    let signals =
        signal_hook::iterator::Signals::new(&[signal_hook::SIGINT, signal_hook::SIGTERM])?;
    let watch_path = std::env::args()
        .nth(1)
        .expect("Failed to read watch path parameter");
    let current_directory = std::env::current_dir().expect("Failed to determine current directory");
    let child_path = if std::path::Path::new(&watch_path).is_absolute() {
        std::path::Path::new(&watch_path).to_path_buf()
    } else {
        current_directory.join(&watch_path)
    };
    std::thread::spawn(move || read_file(&child_path));
    for signal in signals.forever() {
        if signal == signal_hook::SIGINT || signal == signal_hook::SIGTERM {
            break;
        }
    }
    Ok(())
}

# monitor-write

## Status

- `bash`: Only monitor, signals not forwarded (does not stop on signals)
- `golang`: Only monitor
- `rust`: Only monitor, threads not stopped gracefully (no `std::thread::join()`)

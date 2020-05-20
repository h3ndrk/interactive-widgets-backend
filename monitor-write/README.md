# monitor-write

## Status

- `bash`: Only monitor, signals not forwarded (does not stop on signals)
- `rust`: Only monitor, threads not stopped gracefully (no `std::thread::join()`)

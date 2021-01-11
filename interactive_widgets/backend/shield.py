import asyncio


async def shield(awaitable):
    task = asyncio.create_task(awaitable)
    cancelled = False
    while True:
        try:
            await asyncio.shield(task)
        except asyncio.CancelledError:
            cancelled = True
        if task.done():
            break
    if cancelled:
        raise asyncio.CancelledError
    return await task

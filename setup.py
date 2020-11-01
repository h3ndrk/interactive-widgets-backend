import setuptools

setuptools.setup(
    name='inter_md',
    version='0.0.1',
    packages=[
        'inter_md',
    ],
    entry_points={
        'console_scripts': [
            'inter-md = inter_md.backend.main:main',
            'inter-md-monitor = inter_md.monitor.main:main',
        ],
    },
    install_requires=[
        'aiodocker>=0.19.1',
        'aiofiles>=0.6.0',
        'aiohttp>=3.6.2',
        'asyncinotify>=1.0.0',
        'click>=7.1.2',
        'inotify>=0.2.10',
    ],
)

import setuptools

setuptools.setup(
    name='inter_md',
    version='0.0.1',
    packages=setuptools.find_packages(),
    entry_points={
        'console_scripts': [
            'inter-md-backend = inter_md.backend.main:main',
            'inter-md-builder = inter_md.builder.main:main',
            'inter-md-monitor = inter_md.monitor.main:main',
        ],
    },
    install_requires=[
        'aiodocker>=0.19.1',
        'aiofiles>=0.6.0',
        'aiohttp>=3.6.2',
        'asyncinotify>=1.0.0',
        'beautifulsoup4>=4.9.3',
        'click>=7.1.2',
        'html5lib>=1.1',
        'inotify>=0.2.10',
    ],
)

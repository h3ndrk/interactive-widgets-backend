import setuptools

setuptools.setup(
    name='interactive_widgets',
    version='0.0.1',
    packages=setuptools.find_packages(),
    entry_points={
        'console_scripts': [
            'interactive-widgets-backend = interactive_widgets.backend.main:main',
            'interactive-widgets-monitor = interactive_widgets.monitor.main:main',
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

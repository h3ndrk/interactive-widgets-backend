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
        'aiohttp>=3.6.2',
        'click>=7.1.2',
        'inotify>=0.2.10',
    ],
)

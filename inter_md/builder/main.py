import binascii
import bs4
import click
import hashlib
import json
import logging
import os
import pathlib
import re
import shlex
import subprocess
import typing
import uuid


class Page:

    def __init__(self, arguments: dict, url: pathlib.PurePosixPath, contents: bytes):
        self.arguments = arguments
        self.url = url
        assert os.sep == '/', 'Assumption: posix paths'
        self.soup = bs4.BeautifulSoup(contents, 'html5lib')
        self.required_files: typing.List[pathlib.PurePosixPath] = []
        self.backend_type = arguments['backend_type']
        self.backend_configuration = {
            'type': self.backend_type,
            'logger_name_page': 'Page',
            'logger_name_room_connection': 'RoomConnection',
            'logger_name_room': f'{self.backend_type.capitalize()}Room',
            'executors': {},
        }
        self.add_room_connection()
        self.add_meta_tags()
        self.add_title()
        self.replace_widgets()

    def relative(self, file: pathlib.PurePosixPath):
        return pathlib.PurePosixPath(os.path.relpath(file, self.url))

    def hash_inputs(self, *args):
        return hashlib.sha256(b'-'.join([
            binascii.hexlify(arg.encode("utf-8"))
            for arg in args
        ])).hexdigest()

    def add_room_connection(self):
        script_body = self.soup.new_tag('script')
        script_body.append(
            'const roomConnection = new RoomConnection();',
        )
        self.soup.body.insert(0, script_body)
        script_head = self.soup.new_tag('script')
        script_head['src'] = self.relative('/RoomConnection.js')
        self.soup.head.append(script_head)
        self.required_files.append(self.relative('/RoomConnection.js'))

    def add_meta_tags(self):
        meta_charset = self.soup.new_tag('meta')
        meta_charset['charset'] = 'utf-8'
        self.soup.head.append(meta_charset)
        meta_viewport = self.soup.new_tag('meta')
        meta_viewport['name'] = 'viewport'
        meta_viewport['content'] = 'width=device-width, initial-scale=1.0'
        self.soup.head.append(meta_viewport)

    def add_title(self):
        page_title = self.soup.find('h1')
        assert page_title is not None
        title = self.soup.new_tag('title')
        title.append(page_title.text)
        self.soup.head.append(title)

    def sanitize_javascript_input(self, data: str):
        return re.sub(r'[^0-9a-zA-Z\.\-]', lambda match: f'\\u{{{hex(ord(match.group(0)))[2:]}}}', data)

    def replace_widgets(self):
        for index, widget_element in enumerate(self.soup.find_all('x-button')):
            self.replace_button(index, widget_element)
        for index, widget_element in enumerate(self.soup.find_all('x-image-viewer')):
            self.replace_image_viewer(index, widget_element)
        for index, widget_element in enumerate(self.soup.find_all('x-terminal')):
            self.replace_terminal(index, widget_element)
        for index, widget_element in enumerate(self.soup.find_all('x-text-editor')):
            self.replace_text_editor(index, widget_element)
        for index, widget_element in enumerate(self.soup.find_all('x-text-viewer')):
            self.replace_text_viewer(index, widget_element)

    def replace_button(self, index: int, widget_element):
        name = widget_element['name'] if widget_element.has_attr(
            'name') else self.hash_inputs('button', str(index), widget_element['command'], widget_element.text)
        assert re.fullmatch(r'[0-9a-z\-]+', name) is not None
        command = widget_element['command']
        image = widget_element['image']
        label = widget_element.text

        new_element = self.soup.new_tag('div')
        new_element['id'] = f'widget-button-{name}'
        widget_element.replace_with(new_element)

        script_body = self.soup.new_tag('script')
        script_body.append(
            f'roomConnection.subscribe(new ButtonWidget(document.getElementById("widget-button-{name}"), roomConnection.getSendMessageCallback("{name}"), "{self.sanitize_javascript_input(command)}", "{self.sanitize_javascript_input(label)}"));',
        )
        self.soup.body.append(script_body)

        if self.relative('/ButtonWidget.js') not in self.required_files:
            script_head = self.soup.new_tag('script')
            script_head['src'] = self.relative('/ButtonWidget.js')
            self.soup.head.append(script_head)
            self.required_files.append(self.relative('/ButtonWidget.js'))

        self.backend_configuration['executors'][name] = {
            'type': 'once',
            'logger_name': f'{self.backend_type.capitalize()}Once',
            'image': image,
            'command': shlex.split(command),
        }

    def replace_image_viewer(self, index: int, widget_element):
        name = widget_element['name'] if widget_element.has_attr(
            'name') else self.hash_inputs('image_viewer', str(index), widget_element['file'], widget_element['mime'])
        assert re.fullmatch(r'[0-9a-z\-]+', name) is not None
        file = widget_element['file']
        mime = widget_element['mime']

        new_element = self.soup.new_tag('div')
        new_element['id'] = f'widget-image-viewer-{name}'
        widget_element.replace_with(new_element)

        script_body = self.soup.new_tag('script')
        script_body.append(
            f'roomConnection.subscribe(new ImageViewerWidget(document.getElementById("widget-image-viewer-{name}"), roomConnection.getSendMessageCallback("{name}"), "{self.sanitize_javascript_input(file)}", "{self.sanitize_javascript_input(mime)}"));',
        )
        self.soup.body.append(script_body)

        if self.relative('/ImageViewerWidget.js') not in self.required_files:
            script_head = self.soup.new_tag('script')
            script_head['src'] = self.relative('/ImageViewerWidget.js')
            self.soup.head.append(script_head)
            self.required_files.append(self.relative('/ImageViewerWidget.js'))

        self.backend_configuration['executors'][name] = {
            'type': 'always',
            'logger_name': f'{self.backend_type.capitalize()}Always',
            'image': 'inter-md-monitor',
            'enable_tty': True,
            'command': ['inter-md-monitor', file, '0.1', '5.0'],
        }

    def replace_terminal(self, index: int, widget_element):
        name = widget_element['name'] if widget_element.has_attr(
            'name') else self.hash_inputs('terminal', str(index), widget_element['working-directory'])
        assert re.fullmatch(r'[0-9a-z\-]+', name) is not None
        image = widget_element['image']  # TODO: show in frontend?
        command = widget_element['command']  # TODO: show in frontend?
        working_directory = widget_element['working-directory']

        new_element = self.soup.new_tag('div')
        new_element['id'] = f'widget-terminal-{name}'
        widget_element.replace_with(new_element)

        script_body = self.soup.new_tag('script')
        script_body.append(
            f'roomConnection.subscribe(new TerminalWidget(document.getElementById("widget-terminal-{name}"), roomConnection.getSendMessageCallback("{name}"), "{self.sanitize_javascript_input(working_directory)}"));',
        )
        self.soup.body.append(script_body)

        if self.relative('/TerminalWidget.js') not in self.required_files:
            script_head = self.soup.new_tag('script')
            script_head['src'] = self.relative('/TerminalWidget.js')
            self.soup.head.append(script_head)
            self.required_files.append(self.relative('/TerminalWidget.js'))

            script_head_xterm = self.soup.new_tag('script')
            script_head_xterm['src'] = self.relative(
                '//node_modules/xterm/lib/xterm.js',
            )
            self.soup.head.append(script_head_xterm)
            self.required_files.append(self.relative(
                '//node_modules/xterm/lib/xterm.js',
            ))

            script_head_xterm_fit = self.soup.new_tag('script')
            script_head_xterm_fit['src'] = self.relative(
                '/node_modules/xterm-addon-fit/lib/xterm-addon-fit.js',
            )
            self.soup.head.append(script_head_xterm_fit)
            self.required_files.append(self.relative(
                '/node_modules/xterm-addon-fit/lib/xterm-addon-fit.js',
            ))

            style_head_xterm = self.soup.new_tag('link')
            style_head_xterm['rel'] = 'stylesheet'
            style_head_xterm['href'] = self.relative(
                '/node_modules/xterm/css/xterm.css',
            )
            self.soup.head.append(style_head_xterm)
            self.required_files.append(self.relative(
                '/node_modules/xterm/css/xterm.css',
            ))

        self.backend_configuration['executors'][name] = {
            'type': 'always',
            'logger_name': f'{self.backend_type.capitalize()}Always',
            'image': image,
            'enable_tty': True,
            'command': shlex.split(command),
        }

    def replace_text_editor(self, index: int, widget_element):
        name = widget_element['name'] if widget_element.has_attr(
            'name') else self.hash_inputs('text_editor', str(index), widget_element['file'])
        assert re.fullmatch(r'[0-9a-z\-]+', name) is not None
        file = widget_element['file']

        new_element = self.soup.new_tag('div')
        new_element['id'] = f'widget-text-editor-{name}'
        widget_element.replace_with(new_element)

        script_body = self.soup.new_tag('script')
        script_body.append(
            f'roomConnection.subscribe(new TextEditorWidget(document.getElementById("widget-text-editor-{name}"), roomConnection.getSendMessageCallback("{name}"), "{self.sanitize_javascript_input(file)}"));',
        )
        self.soup.body.append(script_body)

        if self.relative('/TextEditorWidget.js') not in self.required_files:
            script_head = self.soup.new_tag('script')
            script_head['src'] = self.relative('/TextEditorWidget.js')
            self.soup.head.append(script_head)
            self.required_files.append(self.relative('/TextEditorWidget.js'))

            script_head_codemirror = self.soup.new_tag('script')
            script_head_codemirror['src'] = self.relative(
                '/node_modules/codemirror/lib/codemirror.js',
            )
            self.soup.head.append(script_head_codemirror)
            self.required_files.append(self.relative(
                '/node_modules/codemirror/lib/codemirror.js',
            ))

            style_head_codemirror = self.soup.new_tag('link')
            style_head_codemirror['rel'] = 'stylesheet'
            style_head_codemirror['href'] = self.relative(
                '/node_modules/codemirror/lib/codemirror.css',
            )
            self.soup.head.append(style_head_codemirror)
            self.required_files.append(self.relative(
                '/node_modules/codemirror/lib/codemirror.css',
            ))

        self.backend_configuration['executors'][name] = {
            'type': 'always',
            'logger_name': f'{self.backend_type.capitalize()}Always',
            'image': 'inter-md-monitor',
            'enable_tty': True,
            'command': ['inter-md-monitor', file, '0.1', '5.0'],
        }

    def replace_text_viewer(self, index: int, widget_element):
        name = widget_element['name'] if widget_element.has_attr(
            'name') else self.hash_inputs('text_viewer', str(index), widget_element['file'])
        assert re.fullmatch(r'[0-9a-z\-]+', name) is not None
        file = widget_element['file']

        new_element = self.soup.new_tag('div')
        new_element['id'] = f'widget-text-viewer-{name}'
        widget_element.replace_with(new_element)

        script_body = self.soup.new_tag('script')
        script_body.append(
            f'roomConnection.subscribe(new TextViewerWidget(document.getElementById("widget-text-viewer-{name}"), roomConnection.getSendMessageCallback("{name}"), "{self.sanitize_javascript_input(file)}"));',
        )
        self.soup.body.append(script_body)

        if self.relative('/TextViewerWidget.js') not in self.required_files:
            script_head = self.soup.new_tag('script')
            script_head['src'] = self.relative('/TextViewerWidget.js')
            self.soup.head.append(script_head)
            self.required_files.append(self.relative('/TextViewerWidget.js'))

        self.backend_configuration['executors'][name] = {
            'type': 'always',
            'logger_name': f'{self.backend_type.capitalize()}Always',
            'image': 'inter-md-monitor',
            'enable_tty': True,
            'command': ['inter-md-monitor', file, '0.1', '5.0'],
        }


@click.command()
@click.option('--logging-level', default='DEBUG', type=click.Choice(['DEBUG', 'INFO', 'WARNING', 'ERROR', 'CRITICAL']), show_default=True)
@click.option('--backend-host', default='*', show_default=True)
@click.option('--backend-port', default=8080, show_default=True)
@click.option('--backend-logging-level', default='DEBUG', type=click.Choice(['DEBUG', 'INFO', 'WARNING', 'ERROR', 'CRITICAL']), show_default=True)
@click.option('--backend-type', default='docker', type=click.Choice(['docker']), show_default=True)
@click.argument('pages_directory', type=click.Path(exists=True, file_okay=False, dir_okay=True, readable=True))
def main(**arguments):
    logging.basicConfig(
        level=arguments['logging_level'],
        format='%(asctime)s  %(name)-20s  %(levelname)-8s  %(message)s',
    )
    logger = logging.getLogger('main')
    pages_directory = pathlib.Path(arguments['pages_directory'])
    page_file = pages_directory / 'page.md'
    logger.info(f'Hello World: {page_file}')
    process = subprocess.Popen(
        args=['pandoc', '--from', 'markdown', '--to', 'html5', page_file], stdout=subprocess.PIPE)
    stdout, _ = process.communicate()
    print(stdout)
    page = Page(arguments, pathlib.PurePosixPath('/test/foo/bar'), stdout)
    print(page.soup.prettify())
    print(page.url, page.url.as_uri())
    backend_configuration = {
        'host': arguments['backend_host'],
        'port': arguments['backend_port'],
        'logger_name': 'Server',
        'logging_level': arguments['backend_logging_level'],
        'context': {
            'type': page.backend_type,
            'logger_name': f'{page.backend_type.capitalize()}Context',
        },
        'pages': {
            str(page.url): page.backend_configuration,
        },
    }
    import pprint
    pprint.pprint(backend_configuration)
    # print(page.soup.encode(formatter='html5'))
    # soup = bs4.BeautifulSoup(stdout, 'html5lib')
    # print(soup.find_all('x-terminal'))
    # with open(page_file) as f:
    #     markdown = marko.Markdown(extensions=[GitHubWiki])
    #     print(markdown(f.read()))

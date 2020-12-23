import bs4
import click
import json
import logging
import marko
import pathlib
import re
import subprocess
import uuid


class Page:

    def __init__(self, contents: bytes):
        self.soup = bs4.BeautifulSoup(contents, 'html5lib')
        self.setup_widgets_in_head = []
        self.required_files = []
        self.add_room_connection()
        self.add_meta_tags()
        self.add_title()
        self.replace_widgets()

    def add_room_connection(self):
        script_body = self.soup.new_tag('script')
        script_body.append('const roomConnection = new RoomConnection();')
        self.soup.body.insert(0, script_body)
        script_head = self.soup.new_tag('script')
        script_head['src'] = 'RoomConnection.js'
        self.soup.head.append(script_head)
        self.required_files.append('RoomConnection.js')

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
        for widget_element in self.soup.find_all('x-button'):
            self.replace_button(widget_element)
        for widget_element in self.soup.find_all('x-image-viewer'):
            self.replace_image_viewer(widget_element)
        for widget_element in self.soup.find_all('x-terminal'):
            self.replace_terminal(widget_element)
        for widget_element in self.soup.find_all('x-text-editor'):
            self.replace_text_editor(widget_element)
        for widget_element in self.soup.find_all('x-text-viewer'):
            self.replace_text_viewer(widget_element)

    def replace_button(self, widget_element):
        name = widget_element['name'] if widget_element.has_attr(
            'name') else str(uuid.uuid4())
        assert re.fullmatch(r'[0-9a-z\-]+', name) is not None
        command = self.sanitize_javascript_input(widget_element['command'])
        label = self.sanitize_javascript_input(widget_element.text)
        new_element = self.soup.new_tag('div')
        new_element['id'] = f'widget-button--{name}'
        widget_element.replace_with(new_element)
        script_body = self.soup.new_tag('script')
        script_body.append(
            f'roomConnection.subscribe(new ButtonWidget(document.getElementById("widget-button--{name}"), roomConnection.getSendMessageCallback("{name}"), "{command}", "{label}"));',
        )
        self.soup.body.append(script_body)
        if 'ButtonWidget.js' not in self.setup_widgets_in_head:
            self.setup_widgets_in_head.append('ButtonWidget.js')
            script_head = self.soup.new_tag('script')
            script_head['src'] = 'ButtonWidget.js'
            self.soup.head.append(script_head)
            self.required_files.append('ButtonWidget.js')

    def replace_image_viewer(self, widget_element):
        name = widget_element['name'] if widget_element.has_attr(
            'name') else str(uuid.uuid4())
        assert re.fullmatch(r'[0-9a-z\-]+', name) is not None
        file = self.sanitize_javascript_input(widget_element['file'])
        mime = self.sanitize_javascript_input(widget_element['mime'])
        new_element = self.soup.new_tag('div')
        new_element['id'] = f'widget-image-viewer--{name}'
        widget_element.replace_with(new_element)
        script_body = self.soup.new_tag('script')
        script_body.append(
            f'roomConnection.subscribe(new ImageViewerWidget(document.getElementById("widget-image-viewer--{name}"), roomConnection.getSendMessageCallback("{name}"), "{file}", "{mime}"));',
        )
        self.soup.body.append(script_body)
        if 'ImageViewerWidget.js' not in self.setup_widgets_in_head:
            self.setup_widgets_in_head.append('ImageViewerWidget.js')
            script_head = self.soup.new_tag('script')
            script_head['src'] = 'ImageViewerWidget.js'
            self.soup.head.append(script_head)
            self.required_files.append('ImageViewerWidget.js')

    def replace_terminal(self, widget_element):
        name = widget_element['name'] if widget_element.has_attr(
            'name') else str(uuid.uuid4())
        assert re.fullmatch(r'[0-9a-z\-]+', name) is not None
        working_directory = self.sanitize_javascript_input(
            widget_element['working-directory'])
        new_element = self.soup.new_tag('div')
        new_element['id'] = f'widget-terminal--{name}'
        widget_element.replace_with(new_element)
        script_body = self.soup.new_tag('script')
        script_body.append(
            f'roomConnection.subscribe(new TerminalWidget(document.getElementById("widget-terminal--{name}"), roomConnection.getSendMessageCallback("{name}"), "{working_directory}"));',
        )
        self.soup.body.append(script_body)
        if 'TerminalWidget.js' not in self.setup_widgets_in_head:
            self.setup_widgets_in_head.append('TerminalWidget.js')
            script_head = self.soup.new_tag('script')
            script_head['src'] = 'TerminalWidget.js'
            self.soup.head.append(script_head)
            self.required_files.append('TerminalWidget.js')

    def replace_text_editor(self, widget_element):
        name = widget_element['name'] if widget_element.has_attr(
            'name') else str(uuid.uuid4())
        assert re.fullmatch(r'[0-9a-z\-]+', name) is not None
        file = self.sanitize_javascript_input(widget_element['file'])
        new_element = self.soup.new_tag('div')
        new_element['id'] = f'widget-text-editor--{name}'
        widget_element.replace_with(new_element)
        script_body = self.soup.new_tag('script')
        script_body.append(
            f'roomConnection.subscribe(new TextEditorWidget(document.getElementById("widget-text-editor--{name}"), roomConnection.getSendMessageCallback("{name}"), "{file}"));',
        )
        self.soup.body.append(script_body)
        if 'TextEditorWidget.js' not in self.setup_widgets_in_head:
            self.setup_widgets_in_head.append('TextEditorWidget.js')
            script_head = self.soup.new_tag('script')
            script_head['src'] = 'TextEditorWidget.js'
            self.soup.head.append(script_head)
            self.required_files.append('TextEditorWidget.js')

    def replace_text_viewer(self, widget_element):
        name = widget_element['name'] if widget_element.has_attr(
            'name') else str(uuid.uuid4())
        assert re.fullmatch(r'[0-9a-z\-]+', name) is not None
        file = self.sanitize_javascript_input(widget_element['file'])
        new_element = self.soup.new_tag('div')
        new_element['id'] = f'widget-text-viewer--{name}'
        widget_element.replace_with(new_element)
        script_body = self.soup.new_tag('script')
        script_body.append(
            f'roomConnection.subscribe(new TextViewerWidget(document.getElementById("widget-text-viewer--{name}"), roomConnection.getSendMessageCallback("{name}"), "{file}"));',
        )
        self.soup.body.append(script_body)
        if 'TextViewerWidget.js' not in self.setup_widgets_in_head:
            self.setup_widgets_in_head.append('TextViewerWidget.js')
            script_head = self.soup.new_tag('script')
            script_head['src'] = 'TextViewerWidget.js'
            self.soup.head.append(script_head)
            self.required_files.append('TextViewerWidget.js')


@click.command()
@click.option('--logging-level', default='DEBUG', type=click.Choice(['DEBUG', 'INFO', 'WARNING', 'ERROR', 'CRITICAL']), show_default=True)
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
    page = Page(stdout)
    print(page.soup.prettify())
    # print(page.soup.encode(formatter='html5'))
    # soup = bs4.BeautifulSoup(stdout, 'html5lib')
    # print(soup.find_all('x-terminal'))
    # with open(page_file) as f:
    #     markdown = marko.Markdown(extensions=[GitHubWiki])
    #     print(markdown(f.read()))

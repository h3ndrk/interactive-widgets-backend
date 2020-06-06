import React from 'react';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import 'xterm/css/xterm.css';

export default class TerminalWidget extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.terminal = null;
    this.fitAddon = null;
    this.container = null;
    this.state = {
      title: 'Terminal',
    };
  }
  
  componentDidMount() {
    this.terminal = new Terminal({
      fontFamily: 'JetBrains Mono',
      theme: {
        foreground: '#000',
        background: '#eee',
        black: '#000',
        blue: '#2196f3',
        brightBlack: '#000',
        brightBlue: '#2196f3',
        brightCyan: '#00acc1',
        brightGreen: '#43a047',
        brightMagenta: '#9c27b0',
        brightRed: '#f44336',
        brightWhite: '#000',
        brightYellow: '#fdd835',
        cursor: '#000',
        cursorAccent: '#000',
        cyan: '#00acc1',
        green: '#43a047',
        magenta: '#9c27b0',
        red: '#f44336',
        selection: 'rgba(0, 0, 0, 0.25)',
        white: '#000',
        yellow: '#fdd835',
      },
    });
    this.fitAddon = new FitAddon();
    this.terminal.loadAddon(this.fitAddon);
    this.terminal.open(this.container);
    if (this.props.onData) {
      this.terminal.onData(this.props.onData);
    }
    this.terminal.onTitleChange(title => {
      this.setState({
        title: title,
      });
    });
    this.fitAddon.fit();
  }
  
  componentWillUnmount() {
    if (this.terminal) {
      this.terminal.dispose();
      this.terminal = null;
    }
  }
  
  write(data) {
    if (this.terminal) {
      this.terminal.write(data);
    }
  }
  
  render() {
    return (
      <div className="centered terminal">
        <div className="title">{this.state.title}</div>
        <div ref={ref => (this.container = ref)} className="terminal" />
      </div>
    );
  }
}

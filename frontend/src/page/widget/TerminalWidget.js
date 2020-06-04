import React from 'react';
import { Terminal } from 'xterm';
import 'xterm/css/xterm.css';

export default class TerminalWidget extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.terminal = null;
    this.container = null;
    this.state = {
      title: 'Terminal',
    };
  }
  
  componentDidMount() {
    this.terminal = new Terminal();
    this.terminal.open(this.container);
    if (this.props.onData) {
      this.terminal.onData(this.props.onData);
    }
    this.terminal.onTitleChange(title => {
      this.setState({
        title: title,
      });
    });
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

import React from 'react';
import { Terminal } from 'xterm';
import 'xterm/css/xterm.css';

export default class TerminalWidget extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.xterm = null;
    this.container = null;
  }
  
  componentDidMount() {
    this.xterm = new Terminal();
    this.xterm.open(this.container);
  }
  
  componentWillUnmount() {
    if (this.xterm) {
      this.xterm.dispose();
      this.xterm = null;
      if (this.props.onData) {
        this.xterm.on('data', data => this.props.onData(data));
      }
    }
  }
  
  write(data) {
    if (this.xterm) {
      this.xterm.write(data);
    }
  }
  
  render() {
    return (
      <div className="centered terminal">
        <div ref={ref => (this.container = ref)} className="" />
      </div>
    );
  }
}

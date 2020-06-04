import React from 'react';
import { Terminal } from 'xterm';
import 'xterm/css/xterm.css';

export default class TerminalWidget extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.terminal = null;
    this.container = null;
  }
  
  componentDidMount() {
    this.terminal = new Terminal();
    this.terminal.open(this.container);
    console.log('Initialize terminal ...');
    if (this.props.onData) {
      console.log('Register handler ...');
      this.terminal.onData(this.props.onData);
    }
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
        <div ref={ref => (this.container = ref)} className="" />
      </div>
    );
  }
}

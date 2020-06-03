import React from 'react';

export default function ButtonWidget(props) {
  return (
    <div className="centered button">
      <div className="command">{props.widget.command}</div>
      <button className="button">{props.widget.label}</button>
      <div className="outputs">
        {props.widget.outputs.map(output => <div className={output.origin}>{output.data}</div>)}
      </div>
    </div>
  );
}

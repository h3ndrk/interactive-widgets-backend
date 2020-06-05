import React from 'react';

export default function ButtonWidget(props) {
  return (
    <div className="centered button">
      <button className="button" onClick={() => {
        if (props.onClick) {
          props.onClick();
        }
      }}>{props.widget.label}</button>
      <div className="command">{props.widget.command}</div>
      <div className="outputs">
        {props.widget.outputs.map((output, i) =>
          <div className={output.origin} key={i}>{output.data}</div>
        )}
      </div>
    </div>
  );
}

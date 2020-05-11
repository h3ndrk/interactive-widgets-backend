import React, { useEffect } from 'react';
import logo from './logo.svg';
import './App.css';
import socketIOClient from 'socket.io-client';

function App() {
  useEffect(() => {
    const socket = socketIOClient(':3001');
    socket.on('welcome', message => {
      console.log('Welcome:', message);
    });
  }, []);

  return (
    <div className="App">
      <header className="App-header">
        <img src={logo} className="App-logo" alt="logo" />
        <p>
          Edit <code>src/App.js</code> and save to reload. Yay!
        </p>
        <a
          className="App-link"
          href="https://reactjs.org"
          target="_blank"
          rel="noopener noreferrer"
        >
          Learn React
        </a>
      </header>
    </div>
  );
}

export default App;

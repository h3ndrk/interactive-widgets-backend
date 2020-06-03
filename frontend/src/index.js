import React from 'react';
import ReactDOM from 'react-dom';
import { BrowserRouter as Router } from 'react-router-dom';
import './index.css';
import App from './App';

ReactDOM.render(
  <React.StrictMode>
    <Router>
      <App />
    </Router>
  </React.StrictMode>,
  document.getElementById('root')
);

if ('serviceWorker' in navigator) {
  navigator.serviceWorker.ready
    .then(registration => {
      registration.unregister();
    })
    .catch(error => {
      console.error(error.message);
    });
}

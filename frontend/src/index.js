import React from 'react';
import ReactDOM from 'react-dom';
import { BrowserRouter as Router } from 'react-router-dom';
import './index.css';
import './fonts/Inter-Bold-slnt=0.ttf';
import './fonts/Inter-Regular-slnt=0.ttf';
import './fonts/JetBrainsMono-Bold-Italic.ttf';
import './fonts/JetBrainsMono-Bold.ttf';
import './fonts/JetBrainsMono-Italic.ttf';
import './fonts/JetBrainsMono-Regular.ttf';
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

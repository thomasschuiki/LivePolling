class PollingClient {
  constructor(url) {
    this.url = url || this._getWsUrl();
    this.sock = null;
    this.handlers = {};
    this.connected = false;
  }

  _getWsUrl() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    return `${protocol}//${window.location.host}/ws`;
  }

  connect() {
    return new Promise((resolve, reject) => {
      this.sock = new WebSocket(this.url);

      this.sock.onopen = () => {
        this.connected = true;
        this._trigger('connect');
        resolve();
      };

      this.sock.onclose = (e) => {
        this.connected = false;
        this._trigger('close', e);
      };

      this.sock.onerror = (e) => {
        this._trigger('error', e);
        reject(e);
      };

      this.sock.onmessage = (e) => {
        try {
          const msg = JSON.parse(e.data);
          this._handleMessage(msg);
        } catch (err) {
          console.error('Failed to parse message:', err);
        }
      };
    });
  }

  _handleMessage(msg) {
    switch (msg.type) {
      case 'asking':
        this._trigger('question', { question: msg.question, possibilities: msg.possibilities });
        break;
      case 'voted':
        this._trigger('voted', msg.data);
        break;
      case 'statistics':
        this._trigger('statistics', { text: msg.text, choices: msg.choices, votes: msg.votes, id: msg.id });
        break;
      case 'login':
        this._trigger('login', msg.data);
        break;
      case 'msg':
        this._trigger('message', msg.data);
        break;
      case 'clear':
        this._trigger('clear', msg.data);
        break;
      default:
        this._trigger('message', msg);
    }
  }

  on(event, handler) {
    if (!this.handlers[event]) {
      this.handlers[event] = [];
    }
    this.handlers[event].push(handler);
  }

  off(event, handler) {
    if (!this.handlers[event]) return;
    this.handlers[event] = this.handlers[event].filter(h => h !== handler);
  }

  _trigger(event, data) {
    if (this.handlers[event]) {
      this.handlers[event].forEach(h => h(data));
    }
  }

  send(command, data) {
    if (!this.sock || this.sock.readyState !== WebSocket.OPEN) {
      console.error('WebSocket not connected');
      return;
    }
    const msg = { command, data };
    this.sock.send(JSON.stringify(msg));
  }

  init(type) {
    this.send('init', type);
  }

  vote(choiceIndex) {
    this.send('vote', { votenumber: choiceIndex });
  }

  login(username, password) {
    this.send('login', { adminname: username, adminpass: password });
  }

  sendQuestion(question, possibilities) {
    this.send('question', { question, possibilities });
  }

  requestStatistics() {
    this.send('statistics');
  }

  clear() {
    this.send('clear');
  }

  reset() {
    this.send('reset');
  }

  close() {
    if (this.sock) {
      this.sock.close();
    }
  }
}

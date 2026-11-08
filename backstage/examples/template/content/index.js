'use strict';

const http = require('http');

const PORT = Number(process.env.PORT || 3000);
const NAME = process.env.OPENCHOREO_COMPONENT || '${{ values.name }}';

function greeting(name) {
  return `Hello from ${name}!`;
}

function createServer() {
  return http.createServer((req, res) => {
    if (req.url === '/healthz') {
      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ status: 'ok' }));
      return;
    }
    res.writeHead(200, { 'Content-Type': 'text/plain' });
    res.end(`${greeting(NAME)}\n`);
  });
}

if (require.main === module) {
  const server = createServer();
  server.listen(PORT, () => {
    console.log(`${NAME} listening on :${PORT}`);
  });
}

module.exports = { greeting, createServer };

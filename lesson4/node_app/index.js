const http = require('http');

const PORT = process.env.NODE_PORT || 8002;

const server = http.createServer((req, res) => {
  res.writeHead(200, {'Content-Type': 'text/plain'});
  res.end('Hello from Node.js!\n');
});

server.listen(PORT, '0.0.0.0', () => {
  console.log(`Node.js app listening on port ${PORT}`);
});

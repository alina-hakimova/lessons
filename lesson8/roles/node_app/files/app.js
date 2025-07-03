const http = require('http');

const PORT = 5000;
const SECRET_MESSAGE = process.env.SECRET_MESSAGE || 'no secret provided'; // Читает из переменных окружения

const server = http.createServer((req, res) => {
  res.statusCode = 200;
  res.setHeader('Content-Type', 'text/plain');
  res.end(`Secret message: ${SECRET_MESSAGE}\n`);
});

server.listen(PORT, '0.0.0.0', () => {
  console.log(`Server running on port ${PORT}`);
});

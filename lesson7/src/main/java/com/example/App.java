import com.sun.net.httpserver.HttpExchange;
 import com.sun.net.httpserver.HttpHandler;
 import com.sun.net.httpserver.HttpServer;
 import java.io.IOException;
 import java.io.OutputStream;
 import java.net.InetSocketAddress;
 src/main/java/App.java
 public class App {
 public static void main(String[] args) throws Exception {
 int port = 8000;
 HttpServer server = HttpServer.create(new InetSocketAddress(port), 0);
 server.createContext("/", new HttpHandler() {
 @Override
 public void handle(HttpExchange exchange) throws IOException {
 String response = "Hello from Java!";
 exchange.sendResponseHeaders(200, response.length());
 try (OutputStream os = exchange.getResponseBody()) {
os.write(response.getBytes());
 }
 }
 });
 server.start();
 System.out.println("Server running on port " + port);
 }
 }

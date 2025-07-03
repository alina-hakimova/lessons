public class Main {
    public static void main(String[] args) {
        final String secretMessage = System.getenv("SECRET_MESSAGE");
        final String defaultMessage = "no secret provided";
        final String responseMessage = "Secret message: " + 
            (secretMessage != null ? secretMessage : defaultMessage) + "\n";

        System.out.println("Starting server on port 5000...");
        System.out.println(responseMessage.trim());

        try {
            com.sun.net.httpserver.HttpServer server = com.sun.net.httpserver.HttpServer.create(
                new java.net.InetSocketAddress("0.0.0.0", 5000), 0);
            
            server.createContext("/", exchange -> {
                exchange.sendResponseHeaders(200, responseMessage.length());
                exchange.getResponseBody().write(responseMessage.getBytes());
                exchange.close();
            });
            
            server.start();
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}
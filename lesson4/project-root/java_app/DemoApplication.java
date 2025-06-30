import org.springframework.boot.*;
import org.springframework.boot.autoconfigure.*;
import org.springframework.web.bind.annotation.*;

@RestController
@SpringBootApplication
public class DemoApplication {
    @RequestMapping("/")
    String home() {
        return "Hello from Java!";
    }
    public static void main(String[] args) {
        SpringApplication app = new SpringApplication(DemoApplication.class);
        String port = System.getenv("JAVA_PORT");
        if (port == null) port = "8004";
        app.setDefaultProperties(java.util.Collections.singletonMap("server.port", port));
        app.run(args);
    }
}

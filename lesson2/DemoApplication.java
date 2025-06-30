package com.example.demo;
 import org.springframework.boot.*;
 import org.springframework.boot.autoconfigure.*;
 import org.springframework.web.bind.annotation.*;
 @RestController
 @SpringBootApplication
 public class DemoApplication {
 public static void main(String[] args) {
 SpringApplication.run(DemoApplication.class, args);
 }
 @GetMapping("/")
 public String home() {
 return "Hello from Java Spring Boot!";
 }
 }

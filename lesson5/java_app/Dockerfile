FROM openjdk:17-alpine AS builder
WORKDIR /app
COPY DemoApplication.java .
RUN javac DemoApplication.java && \
    jar cfe app.jar DemoApplication *.class

FROM openjdk:17-alpine
WORKDIR /app
COPY --from=builder /app/app.jar .
EXPOSE 8004
ENTRYPOINT ["java", "-jar", "app.jar"]
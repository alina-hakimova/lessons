FROM eclipse-temurin:21-jdk-alpine AS build
WORKDIR /app
COPY DemoApplication.java .
RUN apk add --no-cache maven && \
    printf "<project><modelVersion>4.0.0</modelVersion><groupId>demo</groupId><artifactId>demo</artifactId><version>1.0</version><dependencies><dependency><groupId>org.springframework.boot</groupId><artifactId>spring-boot-starter-web</artifactId><version>3.3.0</version></dependency></dependencies><build><plugins><plugin><groupId>org.apache.maven.plugins</groupId><artifactId>maven-compiler-plugin</artifactId><version>3.11.0</version><configuration><source>21</source><target>21</target></configuration></plugin><plugin><groupId>org.springframework.boot</groupId><artifactId>spring-boot-maven-plugin</artifactId></plugin></plugins></build></project>" > pom.xml && \
    mvn package -DskipTests

FROM eclipse-temurin:21-jre-alpine
WORKDIR /app
ARG UID
ARG GID
RUN addgroup -g $GID appgroup && adduser -D -u $UID -G appgroup appuser
COPY --from=build /app/target/demo-1.0.jar app.jar
USER appuser
EXPOSE 8004
HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD wget -qO- http://localhost:8004/ || exit 1
CMD ["java", "-jar", "app.jar"]

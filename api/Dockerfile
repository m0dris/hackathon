FROM gradle:8.14-jdk21 AS build

COPY --chown=gradle:gradle build.gradle.kts settings.gradle.kts /home/gradle/src/
COPY --chown=gradle:gradle gradle.lockfile /home/gradle/src/
COPY --chown=gradle:gradle src /home/gradle/src/src
WORKDIR /home/gradle/src

RUN gradle bootJar -x test --no-daemon

FROM eclipse-temurin:21-jre

RUN groupadd --system app && useradd --system --gid app --no-create-home app

WORKDIR /app
COPY --from=build /home/gradle/src/build/libs/*-SNAPSHOT.jar /app/app.jar

USER app

EXPOSE 8080 9090

ENTRYPOINT ["java", "-jar", "/app/app.jar"]

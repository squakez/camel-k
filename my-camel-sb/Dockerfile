FROM registry.access.redhat.com/ubi8/openjdk-17:1.16

COPY --chown=185 target/my-service-1.0.0.jar /deployments/my-camel-app.jar

ENV JAVA_APP_JAR="/deployments/my-camel-app.jar"
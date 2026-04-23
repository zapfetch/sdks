plugins {
    `java-library`
    id("com.vanniktech.maven.publish") version "0.30.0"
}

group = "com.zapfetch"
version = "0.1.0"

java {
    sourceCompatibility = JavaVersion.VERSION_11
    targetCompatibility = JavaVersion.VERSION_11
}

repositories {
    mavenCentral()
}

sourceSets {
    create("examples") {
        java.srcDir("examples/src/main/java")
        compileClasspath += sourceSets["main"].output
        runtimeClasspath += sourceSets["main"].output
    }
}

dependencies {
    api("com.squareup.okhttp3:okhttp:4.12.0")
    api("com.fasterxml.jackson.core:jackson-databind:2.17.2")
    api("com.fasterxml.jackson.core:jackson-annotations:2.17.2")
    api("com.fasterxml.jackson.datatype:jackson-datatype-jdk8:2.17.2")

    testImplementation("org.junit.jupiter:junit-jupiter:5.10.3")
    testRuntimeOnly("org.junit.platform:junit-platform-launcher:1.10.3")

    "examplesImplementation"(sourceSets["main"].output)
    "examplesImplementation"("com.fasterxml.jackson.core:jackson-annotations:2.17.2")
}

tasks.test {
    useJUnitPlatform()
}

tasks.withType<Javadoc> {
    options {
        (this as StandardJavadocDocletOptions).apply {
            addStringOption("Xdoclint:none", "-quiet")
        }
    }
}

mavenPublishing {
    publishToMavenCentral(com.vanniktech.maven.publish.SonatypeHost.CENTRAL_PORTAL)
    signAllPublications()

    coordinates("com.zapfetch", "zapfetch-java", version.toString())

    pom {
        name.set("ZapFetch Java SDK")
        description.set("Official Java SDK for the ZapFetch v2 web scraping API.")
        url.set("https://github.com/zapfetch/sdks")

        licenses {
            license {
                name.set("MIT License")
                url.set("https://opensource.org/licenses/MIT")
            }
        }

        developers {
            developer {
                name.set("ZapFetch")
                url.set("https://zapfetch.com")
            }
        }

        scm {
            url.set("https://github.com/zapfetch/sdks")
            connection.set("scm:git:git://github.com/zapfetch/sdks.git")
            developerConnection.set("scm:git:ssh://github.com/zapfetch/sdks.git")
        }
    }
}

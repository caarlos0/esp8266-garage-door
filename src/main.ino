#define closes 15    // D8
#define opens 13     // D7
#define sensor 14    // D5

#include <ESP8266WiFi.h>
#include <WiFiClient.h>
#include <ESP8266WebServer.h>
#include "arduino_secrets.h"

const char* ssid = WIFI_SSID;
const char* password = WIFI_PASS;

ESP8266WebServer server(80);
WiFiClient net;
String lastAction = "none";

void setup() {
  pinMode(closes, OUTPUT);
  pinMode(opens, OUTPUT);
  pinMode(sensor, INPUT_PULLUP);
  pinMode(LED_BUILTIN, OUTPUT);
  Serial.begin(115200);

  WiFi.mode(WIFI_STA);
  WiFi.begin(ssid, password);
  Serial.println("");

  while (WiFi.status() != WL_CONNECTED) {
    delay(100);
    Serial.print(".");
  }
  Serial.println("");
  Serial.print("Connected to ");
  Serial.println(ssid);
  Serial.print("IP address: ");
  Serial.println(WiFi.localIP());

  server.on("/", handleRoot);
  server.on("/open", handleOpen);
  server.on("/close", handleClose);
  server.on("/sensor", handleSensor);
  server.begin();
  Serial.println("HTTP server started");
  digitalWrite(LED_BUILTIN, HIGH);
}

void loop() {
  server.handleClient();
}

void handleRoot() {
  digitalWrite(LED_BUILTIN, LOW);
  String body = "<html>";
  body += "<head><title>Gate Status</title><head>";
  body += "<body>";
  body += "<h1>Gate ESP8266 Status</h1>\n";
  body += "<p>Sensor status: <b>";
  body += digitalRead(sensor) ? "open" : "closed";
  body += "</b></p>\n";
  body += "<p>Last action: <b>"+lastAction+"</b></p>\n";
  body += "<p>";
  body += "<a href=\"/open\">Open</a>";
  body += " | ";
  body += "<a href=\"/close\">Close</a>";
  body += "</p>\n";
  body += "</body>";
  body += "</html>";
  server.send(200, "text/html", body);
  digitalWrite(LED_BUILTIN, HIGH);
}

void handleSensor() {
  server.send(200, "text/html", digitalRead(sensor) ? "open" : "closed");
}


void handleOpen() {
  Serial.println("opening");
  digitalWrite(LED_BUILTIN, LOW);
  digitalWrite(opens, HIGH);
  delay(500);
  digitalWrite(opens, LOW);
  lastAction = "open";
  digitalWrite(LED_BUILTIN, HIGH);
  server.sendHeader("Location","/");
  server.send(303);
}

void handleClose() {
  Serial.println("closing");
  digitalWrite(LED_BUILTIN, LOW);
  digitalWrite(closes, HIGH);
  delay(500);
  digitalWrite(closes, LOW);
  lastAction = "close";
  digitalWrite(LED_BUILTIN, HIGH);
  server.sendHeader("Location","/");
  server.send(303);
}

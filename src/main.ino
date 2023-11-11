#define CLOSES 15    // D8
#define OPENS 13     // D7
#define SENSOR 12    // D6
#define PULSE 500

#include <ESP8266WiFi.h>
#include <WiFiClient.h>
#include <PubSubClient.h>
#include "arduino_secrets.h"

const char* ssid = WIFI_SSID;
const char* password = WIFI_PASS;
const char* mqttHost = MQTT_HOST;

WiFiClient net;
PubSubClient client(net);
String lastStatus = "unknown";

void setup() {
  pinMode(CLOSES, OUTPUT);
  pinMode(OPENS, OUTPUT);
  pinMode(SENSOR, INPUT_PULLUP);
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

  client.setServer(mqttHost, 1883);
  client.setCallback(callback);
  digitalWrite(LED_BUILTIN, HIGH);
}

void callback(char* topic, byte* payload, unsigned int length) {
  Serial.print("got message: ");
  for (int i = 0; i < length; i++) {
    Serial.print((char)payload[i]);
  }
  Serial.println();

  if ((char)payload[0] == 'p') { // ping
    lastStatus = "unknown";
    pubSensor();
  } else if ((char)payload[0] == 'o') { // open
    handleOpen();
  } else { // close
    handleClose();
  }
}

void reconnect() {
  while (!client.connected()) {
    if (client.connect("espgate")) {
      Serial.println("connected");
      client.subscribe("espgate/act");
      lastStatus = "unknown";
      pubSensor();
    } else {
      Serial.print("failed, rc=");
      Serial.print(client.state());
      Serial.println(" try again in 5 seconds");
      delay(5000);
    }
  }
}


void loop() {
  if (!client.connected()) {
    reconnect();
  }
  pubSensor();
  client.loop();
  delay(100);
}

void pubSensor() {
  String status = digitalRead(SENSOR) ? "open" : "closed";
  if (status != lastStatus) {
    client.publish("espgate/sensor", status);
    lastStatus = status;
  }
}

void handleOpen() {
  Serial.println("opening");
  digitalWrite(LED_BUILTIN, LOW);
  digitalWrite(OPENS, HIGH);
  delay(PULSE);
  digitalWrite(OPENS, LOW);
  digitalWrite(LED_BUILTIN, HIGH);
}

void handleClose() {
  Serial.println("closing");
  digitalWrite(LED_BUILTIN, LOW);
  digitalWrite(CLOSES, HIGH);
  delay(PULSE);
  digitalWrite(CLOSES, LOW);
  digitalWrite(LED_BUILTIN, HIGH);
}

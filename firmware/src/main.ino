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
const String hostname = "espgate";

WiFiClient net;
PubSubClient client(net);
String lastStatus = "unknown";

void setup() {
  pinMode(CLOSES, OUTPUT);
  pinMode(OPENS, OUTPUT);
  digitalWrite(OPENS, HIGH);
  digitalWrite(CLOSES, HIGH);
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
  digitalWrite(LED_BUILTIN, LOW);
  if ((char)payload[0] == 'p') { // ping
    Serial.println("ping");
    lastStatus = "unknown";
    pubSensor();
  } else if ((char)payload[0] == 'o') { // open
    Serial.println("open");
    digitalWrite(OPENS, LOW);
  } else { // close
    Serial.println("close");
    digitalWrite(CLOSES, LOW);
  }
  delay(PULSE);
  digitalWrite(OPENS, HIGH);
  digitalWrite(CLOSES, HIGH);
  digitalWrite(LED_BUILTIN, HIGH);
}

void reconnect() {
  while (!client.connected()) {
    if (client.connect(hostname.c_str())) {
      Serial.println("connected");
      client.subscribe((hostname + "/act").c_str());
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
  // NC, change to NO
  String status = digitalRead(SENSOR) ? "open" : "closed";
  if (status != lastStatus) {
    client.publish((hostname+"/sensor").c_str(), status.c_str());
    lastStatus = status;
  }
}

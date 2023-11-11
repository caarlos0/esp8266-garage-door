#include "arduino_secrets.h"
#include <ESP8266WiFi.h>
#include <WiFiClient.h>

const char *ssid = WIFI_SSID;
const char *password = WIFI_PASS;

void wifi_connect() {
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
}

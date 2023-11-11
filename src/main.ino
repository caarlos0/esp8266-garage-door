#define closes 15    // D8
#define opens 13     // D7
#define sensor 14    // D5

#define LOG_D(fmt, ...)   printf_P(PSTR(fmt "\n") , ##__VA_ARGS__);

extern "C" homekit_server_config_t config;

const pulse = 500;
const movementTime = 30000;

#include <Arduino.h>
#include <arduino_homekit_server.h>
#include "wifi.h"

#define LOG_D(fmt, ...)   printf_P(PSTR(fmt "\n") , ##__VA_ARGS__);

#define HOMEKIT_CHARACTERISTIC_CURRENT_DOOR_STATE_OPEN 0
#define HOMEKIT_CHARACTERISTIC_CURRENT_DOOR_STATE_CLOSED 1
#define HOMEKIT_CHARACTERISTIC_CURRENT_DOOR_STATE_OPENING 2
#define HOMEKIT_CHARACTERISTIC_CURRENT_DOOR_STATE_CLOSING 3
#define HOMEKIT_CHARACTERISTIC_CURRENT_DOOR_STATE_STOPPED 4
#define HOMEKIT_CHARACTERISTIC_CURRENT_DOOR_STATE_UNKNOWN 255

#define HOMEKIT_CHARACTERISTIC_TARGET_DOOR_STATE_OPEN 0
#define HOMEKIT_CHARACTERISTIC_TARGET_DOOR_STATE_CLOSED 1
#define HOMEKIT_CHARACTERISTIC_TARGET_DOOR_STATE_UNKNOWN 255

#define HOMEKIT_CHARACTERISTIC_LOCK_CURRENT_STATE_UNSECURED 0
#define HOMEKIT_CHARACTERISTIC_LOCK_CURRENT_STATE_SECURED 1
#define HOMEKIT_CHARACTERISTIC_LOCK_CURRENT_STATE_JAMMED 2
#define HOMEKIT_CHARACTERISTIC_LOCK_CURRENT_STATE_UNKNOWN 3

#define HOMEKIT_CHARACTERISTIC_LOCK_TARGET_STATE_UNSECURED 0
#define HOMEKIT_CHARACTERISTIC_LOCK_TARGET_STATE_SECURED 1
#define HOMEKIT_CHARACTERISTIC_LOCK_TARGET_STATE_JAMMED 2
#define HOMEKIT_CHARACTERISTIC_LOCK_TARGET_STATE_UNKNOWN 3


extern "C" homekit_server_config_t config;
extern "C" homekit_characteristic_t cha_current_door_state;
extern "C" homekit_characteristic_t cha_target_door_state;
extern "C" homekit_characteristic_t cha_obstruction_detected;
extern "C" homekit_characteristic_t cha_name;
extern "C" homekit_characteristic_t cha_lock_current_state;
extern "C" homekit_characteristic_t cha_lock_target_state;


homekit_value_t cha_current_door_state_getter() {
  homekit_characteristic_t current_state = cha_current_door_state;
  if (digitalRead(sensor)) {
    cha_current_door_state.value.uint8_value = HOMEKIT_CHARACTERISTIC_CURRENT_DOOR_STATE_OPEN;
  } else {
    cha_current_door_state.value.uint8_value = HOMEKIT_CHARACTERISTIC_CURRENT_DOOR_STATE_CLOSED;
  }

  // cha_current_door_state.value.uint8_value = HOMEKIT_CHARACTERISTIC_CURRENT_DOOR_STATE_OPENING;
  // cha_current_door_state.value.uint8_value = HOMEKIT_CHARACTERISTIC_CURRENT_DOOR_STATE_CLOSING;

  homekit_characteristic_notify(&cha_target_door_state, cha_target_door_state.value);
	LOG_D("Current door state: %i", cha_current_door_state.value.uint8_value);
	return cha_current_door_state.value;
}

homekit_value_t cha_obstruction_detected_getter() {
	LOG_D("Obstruction: %s", cha_obstruction_detected.value.bool_value ? "Detected" : "Not detected");
	return cha_obstruction_detected.value;
}

homekit_value_t cha_lock_current_state_getter() {
	LOG_D("Current lock state: %i", cha_lock_current_state.value.uint8_value);
	return cha_lock_current_state.value;
}

void cha_target_door_state_setter(const homekit_value_t value) {
  cha_target_door_state.value = value;
	LOG_D("Target door state: %i", value.uint8_value);
  if (cha_current_door_state.value.uint8_value == HOMEKIT_CHARACTERISTIC_TARGET_DOOR_STATE_OPEN && cha_current_door_state.value.uint8_value != cha_target_door_state.value.uint8_value) {
    digitalWrite(opens, HIGH);
    delay(pulse);
    digitalWrite(opens, LOW);
  }
  if (cha_current_door_state.value.uint8_value == HOMEKIT_CHARACTERISTIC_TARGET_DOOR_STATE_CLOSED && cha_current_door_state.value.uint8_value != cha_target_door_state.value.uint8_value) {
    digitalWrite(closes, HIGH);
    delay(pulse);
    digitalWrite(closes, LOW);
  }

}


void my_homekit_setup() {
	cha_current_door_state.getter = cha_current_door_state_getter;
	cha_target_door_state.setter = cha_target_door_state_setter;
	cha_lock_current_state.getter = cha_lock_current_state_getter;
	cha_obstruction_detected.getter = cha_obstruction_detected_getter;

	arduino_homekit_setup(&config);
}

void setup() {
  pinMode(closes, OUTPUT);
  pinMode(opens, OUTPUT);
  pinMode(sensor, INPUT_PULLUP);


  Serial.begin(115200);
  wifi_connect(); // in wifi.h
  //homekit_storage_reset(); // to remove the previous HomeKit pairing storage when you first run this new HomeKit example
	cha_current_door_state.getter = cha_current_door_state_getter;
	cha_target_door_state.setter = cha_target_door_state_setter;
	cha_lock_current_state.getter = cha_lock_current_state_getter;
	cha_obstruction_detected.getter = cha_obstruction_detected_getter;

	arduino_homekit_setup(&config);

  homekit_value_t current_state = cha_current_door_state_getter();
  homekit_characteristic_notify(&cha_current_door_state, cha_current_door_state.value);

  switch (cha_current_door_state.value.uint8_value) {
    case HOMEKIT_CHARACTERISTIC_CURRENT_DOOR_STATE_CLOSED:
      cha_target_door_state.value.uint8_value = HOMEKIT_CHARACTERISTIC_TARGET_DOOR_STATE_CLOSED;
      break;
    case HOMEKIT_CHARACTERISTIC_CURRENT_DOOR_STATE_CLOSING:
      cha_target_door_state.value.uint8_value = HOMEKIT_CHARACTERISTIC_TARGET_DOOR_STATE_CLOSED;
      break;
    case HOMEKIT_CHARACTERISTIC_CURRENT_DOOR_STATE_OPEN:
      cha_target_door_state.value.uint8_value = HOMEKIT_CHARACTERISTIC_TARGET_DOOR_STATE_OPEN;
      break;
    case HOMEKIT_CHARACTERISTIC_CURRENT_DOOR_STATE_OPENING:
      cha_target_door_state.value.uint8_value = HOMEKIT_CHARACTERISTIC_TARGET_DOOR_STATE_OPEN;
      break;
    default:
      cha_target_door_state.value.uint8_value = HOMEKIT_CHARACTERISTIC_TARGET_DOOR_STATE_UNKNOWN;
      break;
  }
  homekit_characteristic_notify(&cha_target_door_state, cha_target_door_state.value);
}

void loop() {
	arduino_homekit_loop();
  delay(10);
}

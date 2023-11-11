# esp8266-garage-door

This is the second revision of this project, the first one used a shelly, this one doesn't.

---

## Wiring

![image](https://github.com/caarlos0/esp8266-garage-door/assets/245435/45faae45-a243-4044-9b66-1a9328026ccb)

Basically, the firmware publish to the `espgate/sensor` mqtt topic, which the homekit-garage software subscribes.
Homekit-garage publishes `open`/`close` events to the `espgate/act` mqtt topic, and the firware acts accordingly.

The firware is pretty "dumb", on purpose, as its the hardest part to maintain/update.

The homekit controller runs in a raspberry pi and is very lightweight too.

## Configuration

In the firmware, you might want to change the `PULSE` (default 500ms) time.
You'll also need to create an `arduino_secrets.h` with the Wifi info and MQTT server IP.

In the homekit controller, you'll need to set the `operationTimeout` (default 30s) and the MQTT server IP.

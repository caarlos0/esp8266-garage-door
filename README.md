# esp8266-garage-door

This is the third revision of this project.

1. the first revision used a shelly1 and "forwarded" the pulses to an esp8266, which forwarded to the correct remote button
2. the second revision ditched the shelly, and was connected directly to the remote. This version also introduced the Homekit controller and the mqtt api
3. the thrid revision adds a dual relay, so the remote electric circuit is isolated from the esp8266's

---

## Wiring

![image](https://github.com/caarlos0/esp8266-garage-door/assets/245435/ad5000ea-42c2-4b7c-be9f-4c4d9f2810d8)

Basically, the firmware publish to the `espgate/sensor` mqtt topic, which `homekit-garage` subscribes.
`homekit-garage` publishes `open`, `close` and `ping` events to the `espgate/act` mqtt topic, and the firmware acts accordingly.

The firmware is pretty "dumb", on purpose, as its the hardest part to maintain/update.

The Homekit controller runs in a raspberry pi, and is very lightweight too.

## Configuration

In the firmware, you might want to change the `PULSE` (default 500ms) time.
You'll also need to create an `arduino_secrets.h` with the Wifi info and MQTT server IP.

In the Homekit controller, you'll need to set the `operationTimeout` (default 30s) and the MQTT server IP.

## Future

Maybe ditch the separated Go Homekit controller and run it directly in the ESP?
Not sure if worth it, as its painful to update/debug...

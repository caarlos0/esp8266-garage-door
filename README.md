# esp8266-garage-door

This is the firmware for a esp8266 d1 mini to "forward" the pulse of a shelly1 to either one gpio or another, depending on a sensor. 

My garage does not have the BTN inputs in its board, and it also has different buttons for opening and closing, so I'll solder it to one of the remotes directly.

It works like this:

- 1 sensor connected to the shelly (used for homekit status only)
- 1 sensor connected to the esp (used to decide which button to press)
- controller buttons connected to esp
- esp connected to shelly
- for now, I'm using 2 power sources because who has the time to figure all that out

The shelly closes the relay, which is read by the esp, which then reads its sensor, and decide if it should open or close the garage.

The wiring is something like this:

 ![image](https://github.com/caarlos0/esp8266-garage-door/assets/245435/b47e9936-df39-45d8-9466-b71fde070fd8)

PS: a friend who is an eletrical engineer helped me a lot on figuring this all out!

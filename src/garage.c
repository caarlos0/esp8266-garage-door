#include <homekit/characteristics.h>
#include <homekit/homekit.h>

void my_accessory_identify(homekit_value_t _value) {
  printf("accessory identify\n");
}

homekit_characteristic_t cha_current_door_state =
    HOMEKIT_CHARACTERISTIC_(CURRENT_DOOR_STATE, 4);

homekit_characteristic_t cha_target_door_state =
    HOMEKIT_CHARACTERISTIC_(TARGET_DOOR_STATE, NULL);

homekit_characteristic_t cha_obstruction_detected =
    HOMEKIT_CHARACTERISTIC_(OBSTRUCTION_DETECTED, 0);

homekit_characteristic_t cha_name =
    HOMEKIT_CHARACTERISTIC_(NAME, "Internal Gate");

homekit_characteristic_t cha_lock_current_state =
    HOMEKIT_CHARACTERISTIC_(LOCK_CURRENT_STATE, 3);

homekit_characteristic_t cha_lock_target_state =
    HOMEKIT_CHARACTERISTIC_(LOCK_TARGET_STATE, NULL);

homekit_accessory_t *accessories[] = {
    HOMEKIT_ACCESSORY(
            .id = 1, .category = homekit_accessory_category_garage,
            .services =
                (homekit_service_t *[]){
                    HOMEKIT_SERVICE(
                        ACCESSORY_INFORMATION,
                        .characteristics =
                            (homekit_characteristic_t *[]){
                                HOMEKIT_CHARACTERISTIC(NAME, "Internal Gate"),
                                HOMEKIT_CHARACTERISTIC(MANUFACTURER,
                                                       "Becker Software"),
                                HOMEKIT_CHARACTERISTIC(SERIAL_NUMBER, "00"),
                                HOMEKIT_CHARACTERISTIC(MODEL, "ESP8266"),
                                HOMEKIT_CHARACTERISTIC(FIRMWARE_REVISION,
                                                       "0.1.0"),
                                HOMEKIT_CHARACTERISTIC(IDENTIFY,
                                                       my_accessory_identify),
                                NULL}),
                    HOMEKIT_SERVICE(GARAGE_DOOR_OPENER, .primary = true,
                                    .characteristics =
                                        (homekit_characteristic_t *[]){
                                            &cha_current_door_state,
                                            &cha_target_door_state,
                                            &cha_obstruction_detected,
                                            &cha_name, &cha_lock_current_state,
                                            &cha_lock_target_state, NULL}),
                    NULL}),
    NULL};

homekit_server_config_t config = {.accessories = accessories,
                                  .password = "001-02-003"};

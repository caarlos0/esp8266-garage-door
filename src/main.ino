#define closes 15    // D8
#define opens 13     // D7
#define activates 12 // D6
#define sensor 14    // D5

void setup() {
  pinMode(closes, OUTPUT);
  pinMode(opens, OUTPUT);
  pinMode(activates, INPUT_PULLUP);
  pinMode(sensor, INPUT_PULLUP);
  pinMode(LED_BUILTIN, OUTPUT);
  Serial.begin(115200);
}

void loop() {
  digitalWrite(LED_BUILTIN, HIGH);
  while (!digitalRead(activates)) {
    Serial.println("active");
    digitalWrite(LED_BUILTIN, LOW);
    if (digitalRead(sensor)) {
      Serial.println("closing");
      digitalWrite(closes, HIGH);
    } else {
      Serial.println("opening");
      digitalWrite(opens, HIGH);
    }
  }
  digitalWrite(closes, LOW);
  digitalWrite(opens, LOW);
  digitalWrite(LED_BUILTIN, HIGH);
}

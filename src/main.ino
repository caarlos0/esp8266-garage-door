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
  Serial.println(digitalRead(sensor));
  while (!digitalRead(activates)) {
    Serial.println("activates");
    digitalWrite(LED_BUILTIN, HIGH);
    if (digitalRead(sensor)) {
      Serial.println("closes");
      digitalWrite(closes, HIGH);
    } else {
      Serial.println("opens");
      digitalWrite(opens, HIGH);
    }
    delay(100);
  }
  digitalWrite(closes, LOW);
  digitalWrite(opens, LOW);
  digitalWrite(LED_BUILTIN, LOW);
}

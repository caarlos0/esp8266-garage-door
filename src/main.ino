#define closes 15    // D8
#define opens 13     // D7
#define activates 12 // D6

bool flag = false;

void setup() {
  pinMode(closes, OUTPUT);
  pinMode(opens, OUTPUT);
  pinMode(activates, INPUT_PULLUP);

  pinMode(LED_BUILTIN, OUTPUT);
  Serial.begin(115200);
}

void loop() {
  // shelly sends 0V to D6 when active
  if (!digitalRead(activates)) {
    Serial.println("activates");
    digitalWrite(LED_BUILTIN, HIGH);
    flag = !flag;
    if (flag) {
      Serial.println("closes");
      digitalWrite(closes, HIGH);
      delay(300);
      digitalWrite(closes, LOW);
    } else if (!flag) {
      Serial.println("opens");
      digitalWrite(opens, HIGH);
      delay(300);
      digitalWrite(opens, LOW);
    }
  }
  while (!digitalRead(activates)) {
    // wait til activates is over
  }
  delay(500);
  digitalWrite(LED_BUILTIN, LOW);
}

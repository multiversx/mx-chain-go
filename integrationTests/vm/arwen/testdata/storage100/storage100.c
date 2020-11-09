#include "context.h"

void store100() {
  byte i;
  byte key[10] = {'m', 'm', 'm', 'm', 'm', 'm', 'm', 'm', 'm', 0};
  byte data[100];

  for (i = 0; i < 100; i++) {
    data[i] = 'a';
  }

  for (i = 0; i < 10; i++) {
    key[9] = i;
    data[99] = i;
    storageStore(key, 10, data, 100);
  }
}

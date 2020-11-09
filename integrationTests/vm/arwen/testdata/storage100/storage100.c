#include "context.h"

byte key[10] = {};
byte data[100] = {};

void store100() {
	byte i;

	// Fill the key with letters
	for (i = 0; i < 10; i++) {
		key[i] = 'f' + i;
	}

	// Fill the data with letters / characters
	for (i = 0; i < 100; i++) {
		data[i] = 'a' + i;
	}

	// Store
	for (i = 0; i < 10; i++) {
		key[9] = i;
		data[99] = i;
		storageStore(key, 10, data, 100);
	}
}

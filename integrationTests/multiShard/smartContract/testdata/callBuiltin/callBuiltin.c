typedef unsigned char byte;
typedef unsigned int i32;
typedef unsigned long long i64;

i64 int64storageLoad(byte *key, int keyLength);
i32 int64storageStore(byte *key, int keyLength, long long value);
void int64finish(i64 value);
int getArgument(int argumentIndex, byte *argument);
void asyncCall(byte *destination, byte *value, byte *data, int length);

byte callValue[32] = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0};
byte receiver[32] = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0};
byte data[11] = "testfunc@01";

void callBuiltin() {
	byte key[5] = "test1";
	int64storageStore(key, 5, 255);

	getArgument(0, receiver);
	asyncCall(receiver, callValue, data, 11);
}

void callBack() {
	byte key[5] = "test2";
	int64storageStore(key, 5, 254);
}

void testValue1() {
	byte key[5] = "test1";
	i64 test1 = int64storageLoad(key, 5);
	int64finish(test1);
}

void testValue2() {
	byte key[5] = "test2";
	i64 test1 = int64storageLoad(key, 5);
	int64finish(test1);
}

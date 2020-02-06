typedef unsigned char byte;
typedef unsigned int i32;
typedef unsigned long long i64;

void asyncCall(byte *destination, byte *value, byte *data, int length);

byte zero[32] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};
byte firstScAddress[32] = {0, 0, 0, 0, 0, 0, 0, 0, 5, 0, 93, 61, 83, 181, 208, 252, 240, 125, 34, 33, 112, 151, 137, 50, 22, 110, 233, 243, 151, 45, 48, 48};

void doSomething()
{
    asyncCall(firstScAddress, zero, (byte*)"callMe@01", sizeof("callMe@01") - 1);
}

void callBack() {
}

void _main(void)
{
}

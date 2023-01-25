typedef unsigned char byte;

typedef unsigned int bigInt;
int getCallValue(byte *resultOffset);
void finish(byte *data, int length);

bigInt bigIntNew(long long value);
void bigIntGetCallValue(bigInt destination);
void bigIntSetInt64(bigInt destination, long long value);
long long bigIntGetInt64(bigInt reference);
void bigIntFinishUnsigned(bigInt reference);

long long fibonacci(long long n) {
  if (n == 0) return 0;
  if (n == 1) return 1;
  return fibonacci(n-1) + fibonacci(n-2);
}

void _main(void) {
	bigInt bcv = bigIntNew(0);
	bigIntGetCallValue(bcv);
	long long cv = bigIntGetInt64(bcv);
	long long result = 0;

	result = fibonacci(cv);

	bigInt bresult = bigIntNew(result);
	bigIntFinishUnsigned(bresult);
}

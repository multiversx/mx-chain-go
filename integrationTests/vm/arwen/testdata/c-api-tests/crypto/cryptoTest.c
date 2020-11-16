#include "../context.h"
#include "../util.h"

typedef byte HASH256[32];
typedef byte HASH160[20];

const int BlsPublicKeyLength = 96;
const int BlsSignatureLength = 48;
const int Ed25519PublicKeyLength = 32;
const int Ed25519SignatureLength = 64;
const int Secp256k1CompressedPublicKeyLength = 33;
const int Secp256k1UncompressedPublicKeyLength = 65;
const int Secp256k1SignatureLength = 64;

const int LIMIT = 10000;
const int MSG_LENGTH = 100;

void* memset(void *dest, int c, unsigned long n);
void* memset(void *dest, int c, unsigned long n)
{
    int i;
    char *cdest = (char *)dest;
    for (i = 0; i < n; i++)
    {
        cdest[i] = c;
    }
    return dest;
}

void init() 
{
    
}

void sha256Test()
{
    int i;
    HASH256 result = { 1 };

    for (i = 0; i < LIMIT; i++)
    {
        sha256(result, sizeof(HASH256), result);
    }
}

void keccak256Test()
{
    int i;
    HASH256 result = { 1 };

    for (i = 0; i < LIMIT; i++)
    {
        keccak256(result, sizeof(HASH256), result);
    }
}

void ripemd160Test()
{
    int i;
    HASH160 result = { 1 };

    for (i = 0; i < LIMIT; i++)
    {
        ripemd160(result, sizeof(HASH160), result);
    }
}

void verifyBLSTest()
{
    int i;
    byte key[BlsPublicKeyLength] = {1};
    byte sig[BlsSignatureLength] = {2};
    byte msg[MSG_LENGTH] = {3};

    for (i = 0; i < LIMIT; i++)
    {
        verifyBLS(key, msg, MSG_LENGTH, sig);
    }
}

void verifyEd25519Test()
{
    int i;
    byte key[Ed25519PublicKeyLength] = {1};
    byte sig[Ed25519SignatureLength] = {2};
    byte msg[MSG_LENGTH] = {3};

    for (i = 0; i < LIMIT/10; i++)
    {
        verifyEd25519(key, msg, MSG_LENGTH, sig);
    }
}

void verifySecp256k1UncompressedKeyTest()
{
    int i;
    byte key[Secp256k1UncompressedPublicKeyLength] = {1};
    byte sig[Secp256k1SignatureLength] = {2};
    byte msg[MSG_LENGTH] = {3};

    for (int i = 0; i < LIMIT; i++)
    {
        verifySecp256k1(key, Secp256k1UncompressedPublicKeyLength, msg, MSG_LENGTH, sig);
    }
}

void verifySecp256k1CompressedKeyTest()
{
    int i;
    byte key[Secp256k1CompressedPublicKeyLength] = {1};
    byte sig[Secp256k1SignatureLength] = {2};
    byte msg[MSG_LENGTH] = {3};

    for (int i = 0; i < LIMIT; i++)
    {
        verifySecp256k1(key, Secp256k1CompressedPublicKeyLength, msg, MSG_LENGTH, sig);
    }
}

void doNothing() {
}

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

void *memcpy(void *dest, const void *src, unsigned long n);
void *memcpy(void *dest, const void *src, unsigned long n)
{
    char *csrc = (char *)src;
    char *cdest = (char *)dest;

    for (int i = 0; i < n; i++)
    {
        cdest[i] = csrc[i];
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
    byte key[BlsPublicKeyLength] = {217, 252, 244, 255, 1, 41, 110, 111, 166, 208, 181, 81, 50, 9, 172, 190, 241, 118, 98, 215, 66, 229, 23, 33, 193, 166, 177, 213, 162, 23, 133, 181, 203, 57, 100, 9, 233, 20, 161, 66, 225, 103, 96, 222, 87, 160, 15, 14, 31, 86, 218, 46, 33, 184, 105, 139, 26, 236, 237, 186, 149, 155, 101, 25, 57, 90, 253, 129, 91, 112, 212, 120, 44, 81, 13, 225, 172, 182, 64, 177, 136, 36, 138, 209, 67, 147, 217, 96, 207, 4, 103, 105, 16, 140, 248, 5};
    byte sig[BlsSignatureLength] = {42, 35, 229, 124, 111, 19, 105, 97, 112, 164, 38, 139, 110, 247, 241, 202, 84, 129, 39, 237, 236, 245, 85, 7, 129, 134, 253, 78, 242, 95, 207, 163, 163, 228, 54, 2, 64, 223, 140, 147, 104, 177, 100, 211, 93, 190, 188, 22};
    byte msg[32] = {127, 124, 180, 2, 210, 197, 15, 86, 173, 174, 69, 88, 57, 240, 162, 45, 225, 106, 36, 155, 195, 94, 235, 118, 98, 161, 148, 106, 156, 104, 38, 152};

    for (i = 0; i < 100; i++)
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

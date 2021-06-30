#include "../context.h"
#include "../bigInt.h"

typedef byte P224PRIVKEY[28];
typedef byte P256PRIVKEY[32];
typedef byte P384PRIVKEY[48];
typedef byte P521PRIVKEY[66];

typedef byte P224MRESULT[57];
typedef byte P256MRESULT[65];
typedef byte P384MRESULT[97];
typedef byte P521MRESULT[133];

const int p224SMResultLength = 28;
const int p256SMResultLength = 32;
const int p384SMResultLength = 48;
const int p521SMResultLength = 64;
typedef byte P224MCRESULT[29];
typedef byte P256MCRESULT[33];
typedef byte P384MCRESULT[49];
typedef byte P521MCRESULT[67];

const int LIMIT = 100;

void init() 
{
}

// void initialVariablesAndCallsTest()
// {
//     int fieldOrderHandle = bigIntNew(0);
//     int basePointOrderHandle = bigIntNew(0);
//     int eqConstantHandle = bigIntNew(0);
//     int p224Handle = p224Ec();
//     int p224XBasePointHandle = bigIntNew(0);
//     int p224YBasePointHandle = bigIntNew(0);
//     ellipticCurveGetValues(p224Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p224XBasePointHandle,p224YBasePointHandle);
//     int p256Handle = p256Ec();
//     int p256XBasePointHandle = bigIntNew(0);
//     int p256YBasePointHandle = bigIntNew(0);
//     ellipticCurveGetValues(p256Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p256XBasePointHandle,p256YBasePointHandle);
//     int p384Handle = p384Ec();
//     int p384XBasePointHandle = bigIntNew(0);
//     int p384YBasePointHandle = bigIntNew(0);
//     ellipticCurveGetValues(p384Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p384XBasePointHandle,p384YBasePointHandle);
//     int p521Handle = p521Ec();
//     int p521XBasePointHandle = bigIntNew(0);
//     int p521YBasePointHandle = bigIntNew(0);
//     ellipticCurveGetValues(p521Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p521XBasePointHandle,p521YBasePointHandle);
// }

void addEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p224XBasePointHandle,p224YBasePointHandle);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p256XBasePointHandle,p256YBasePointHandle);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p384XBasePointHandle,p384YBasePointHandle);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p521XBasePointHandle,p521YBasePointHandle);

    int i;
    for (i = 0; i < LIMIT/4; i++)
    {
        addEC(p224XBasePointHandle,p224YBasePointHandle,p224Handle,p224XBasePointHandle,p224YBasePointHandle,p224XBasePointHandle,p224YBasePointHandle);
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        addEC(p256XBasePointHandle,p256YBasePointHandle,p256Handle,p256XBasePointHandle,p256YBasePointHandle,p256XBasePointHandle,p256YBasePointHandle);
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        addEC(p384XBasePointHandle,p384YBasePointHandle,p384Handle,p384XBasePointHandle,p384YBasePointHandle,p384XBasePointHandle,p384YBasePointHandle);
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        addEC(p521XBasePointHandle,p521YBasePointHandle,p521Handle,p521XBasePointHandle,p521YBasePointHandle,p521XBasePointHandle,p521YBasePointHandle);
    }
}

void doubleEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p224XBasePointHandle,p224YBasePointHandle);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p256XBasePointHandle,p256YBasePointHandle);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p384XBasePointHandle,p384YBasePointHandle);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p521XBasePointHandle,p521YBasePointHandle);

    int i;
    for (i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p224XBasePointHandle,p224YBasePointHandle,p224Handle,p224XBasePointHandle,p224YBasePointHandle);
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p256XBasePointHandle,p256YBasePointHandle,p256Handle,p256XBasePointHandle,p256YBasePointHandle);
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p384XBasePointHandle,p384YBasePointHandle,p384Handle,p384XBasePointHandle,p384YBasePointHandle);
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p521XBasePointHandle,p521YBasePointHandle,p521Handle,p521XBasePointHandle,p521YBasePointHandle);
    }
}

void isOnCurveEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p224XBasePointHandle,p224YBasePointHandle);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p256XBasePointHandle,p256YBasePointHandle);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p384XBasePointHandle,p384YBasePointHandle);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p521XBasePointHandle,p521YBasePointHandle);

    int i;
    for (i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p224XBasePointHandle,p224YBasePointHandle,p224Handle,p224XBasePointHandle,p224YBasePointHandle);
        isOnCurveEC(p224Handle,p224XBasePointHandle,p224YBasePointHandle);
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p256XBasePointHandle,p256YBasePointHandle,p256Handle,p256XBasePointHandle,p256YBasePointHandle);
        isOnCurveEC(p256Handle,p256XBasePointHandle,p256YBasePointHandle);
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p384XBasePointHandle,p384YBasePointHandle,p384Handle,p384XBasePointHandle,p384YBasePointHandle);
        isOnCurveEC(p384Handle,p384XBasePointHandle,p384YBasePointHandle);
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p521XBasePointHandle,p521YBasePointHandle,p521Handle,p521XBasePointHandle,p521YBasePointHandle);
        isOnCurveEC(p521Handle,p521XBasePointHandle,p521YBasePointHandle);
    }
}

void marshalEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p224XBasePointHandle,p224YBasePointHandle);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p256XBasePointHandle,p256YBasePointHandle);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p384XBasePointHandle,p384YBasePointHandle);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p521XBasePointHandle,p521YBasePointHandle);
    P224MRESULT p224MarshalResult;
    P256MRESULT p256MarshalResult;
    P384MRESULT p384MarshalResult;
    P521MRESULT p521MarshalResult;
    
    int i;
    for (i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p224XBasePointHandle,p224YBasePointHandle,p224Handle,p224XBasePointHandle,p224YBasePointHandle);
        marshalEC(p224XBasePointHandle,p224YBasePointHandle,p224Handle,p224MarshalResult);
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p256XBasePointHandle,p256YBasePointHandle,p256Handle,p256XBasePointHandle,p256YBasePointHandle);
        marshalEC(p256XBasePointHandle,p256YBasePointHandle,p256Handle,p256MarshalResult);
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p384XBasePointHandle,p384YBasePointHandle,p384Handle,p384XBasePointHandle,p384YBasePointHandle);
        marshalEC(p384XBasePointHandle,p384YBasePointHandle,p384Handle,p384MarshalResult);
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p521XBasePointHandle,p521YBasePointHandle,p521Handle,p521XBasePointHandle,p521YBasePointHandle);
        marshalEC(p521XBasePointHandle,p521YBasePointHandle,p521Handle,p521MarshalResult);
    }
}

void unmarshalEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p224XBasePointHandle,p224YBasePointHandle);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p256XBasePointHandle,p256YBasePointHandle);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p384XBasePointHandle,p384YBasePointHandle);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p521XBasePointHandle,p521YBasePointHandle);
    P224MRESULT p224MarshalResult;
    P256MRESULT p256MarshalResult;
    P384MRESULT p384MarshalResult;
    P521MRESULT p521MarshalResult;
    
    int i;
    for (i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p224XBasePointHandle,p224YBasePointHandle,p224Handle,p224XBasePointHandle,p224YBasePointHandle);
        marshalEC(p224XBasePointHandle,p224YBasePointHandle,p224Handle,p224MarshalResult);
        unmarshalEC(p224XBasePointHandle,p224YBasePointHandle,p224Handle,p224MarshalResult,sizeof(P224MRESULT));
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p256XBasePointHandle,p256YBasePointHandle,p256Handle,p256XBasePointHandle,p256YBasePointHandle);
        marshalEC(p256XBasePointHandle,p256YBasePointHandle,p256Handle,p256MarshalResult);
        unmarshalEC(p256XBasePointHandle,p256YBasePointHandle,p256Handle,p256MarshalResult,sizeof(P256MRESULT));
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p384XBasePointHandle,p384YBasePointHandle,p384Handle,p384XBasePointHandle,p384YBasePointHandle);
        marshalEC(p384XBasePointHandle,p384YBasePointHandle,p384Handle,p384MarshalResult);
        unmarshalEC(p384XBasePointHandle,p384YBasePointHandle,p384Handle,p384MarshalResult,sizeof(P384MRESULT));
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p521XBasePointHandle,p521YBasePointHandle,p521Handle,p521XBasePointHandle,p521YBasePointHandle);
        marshalEC(p521XBasePointHandle,p521YBasePointHandle,p521Handle,p521MarshalResult);
        unmarshalEC(p521XBasePointHandle,p521YBasePointHandle,p521Handle,p521MarshalResult,sizeof(P521MRESULT));
    }
}

void marshalCompressedEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p224XBasePointHandle,p224YBasePointHandle);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p256XBasePointHandle,p256YBasePointHandle);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p384XBasePointHandle,p384YBasePointHandle);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p521XBasePointHandle,p521YBasePointHandle);
    P224MCRESULT p224MarshalResult;
    P256MCRESULT p256MarshalResult;
    P384MCRESULT p384MarshalResult;
    P521MCRESULT p521MarshalResult;
    
    int i;
    for (i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p224XBasePointHandle,p224YBasePointHandle,p224Handle,p224XBasePointHandle,p224YBasePointHandle);
        marshalCompressedEC(p224XBasePointHandle,p224YBasePointHandle,p224Handle,p224MarshalResult);
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p256XBasePointHandle,p256YBasePointHandle,p256Handle,p256XBasePointHandle,p256YBasePointHandle);
        marshalCompressedEC(p256XBasePointHandle,p256YBasePointHandle,p256Handle,p256MarshalResult);
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p384XBasePointHandle,p384YBasePointHandle,p384Handle,p384XBasePointHandle,p384YBasePointHandle);
        marshalCompressedEC(p384XBasePointHandle,p384YBasePointHandle,p384Handle,p384MarshalResult);
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p521XBasePointHandle,p521YBasePointHandle,p521Handle,p521XBasePointHandle,p521YBasePointHandle);
        marshalCompressedEC(p521XBasePointHandle,p521YBasePointHandle,p521Handle,p521MarshalResult);
    }
}

void unmarshalCompressedEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p224XBasePointHandle,p224YBasePointHandle);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p256XBasePointHandle,p256YBasePointHandle);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p384XBasePointHandle,p384YBasePointHandle);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p521XBasePointHandle,p521YBasePointHandle);
    P224MCRESULT p224MarshalResult;
    P256MCRESULT p256MarshalResult;
    P384MCRESULT p384MarshalResult;
    P521MCRESULT p521MarshalResult;
    
    int i;
    for (i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p224XBasePointHandle,p224YBasePointHandle,p224Handle,p224XBasePointHandle,p224YBasePointHandle);
        marshalCompressedEC(p224XBasePointHandle,p224YBasePointHandle,p224Handle,p224MarshalResult);
        unmarshalCompressedEC(p224XBasePointHandle,p224YBasePointHandle,p224Handle,p224MarshalResult,sizeof(P224MCRESULT));
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p256XBasePointHandle,p256YBasePointHandle,p256Handle,p256XBasePointHandle,p256YBasePointHandle);
        marshalCompressedEC(p256XBasePointHandle,p256YBasePointHandle,p256Handle,p256MarshalResult);
        unmarshalCompressedEC(p256XBasePointHandle,p256YBasePointHandle,p256Handle,p256MarshalResult,sizeof(P256MCRESULT));
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p384XBasePointHandle,p384YBasePointHandle,p384Handle,p384XBasePointHandle,p384YBasePointHandle);
        marshalCompressedEC(p384XBasePointHandle,p384YBasePointHandle,p384Handle,p384MarshalResult);
        unmarshalCompressedEC(p384XBasePointHandle,p384YBasePointHandle,p384Handle,p384MarshalResult,sizeof(P384MCRESULT));
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p521XBasePointHandle,p521YBasePointHandle,p521Handle,p521XBasePointHandle,p521YBasePointHandle);
        marshalCompressedEC(p521XBasePointHandle,p521YBasePointHandle,p521Handle,p521MarshalResult);
        unmarshalCompressedEC(p521XBasePointHandle,p521YBasePointHandle,p521Handle,p521MarshalResult,sizeof(P521MCRESULT));
    }
}

void generateKeyEcTest()
{
    int xPubKeyHandle = bigIntNew(0);
    int yPubKeyHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p256Handle = p256Ec();
    int p384Handle = p384Ec();
    int p521Handle = p521Ec();
    int i;
    P224PRIVKEY p224PrivKey;
    P256PRIVKEY p256PrivKey;
    P384PRIVKEY p384PrivKey;
    P521PRIVKEY p521PrivKey;

    for (i = 0; i < LIMIT/4; i++)
    {
        generateKeyEC(xPubKeyHandle,yPubKeyHandle,p224Handle,p224PrivKey);
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        generateKeyEC(xPubKeyHandle,yPubKeyHandle,p256Handle,p256PrivKey);
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        generateKeyEC(xPubKeyHandle,yPubKeyHandle,p384Handle,p384PrivKey);
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        generateKeyEC(xPubKeyHandle,yPubKeyHandle,p521Handle,p521PrivKey);
    }
 }

void scalarMultEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p224XBasePointHandle,p224YBasePointHandle);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p256XBasePointHandle,p256YBasePointHandle);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p384XBasePointHandle,p384YBasePointHandle);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle,fieldOrderHandle,basePointOrderHandle,eqConstantHandle,p521XBasePointHandle,p521YBasePointHandle);

    byte p224Scalar[p224SMResultLength] = { 0x7f, 0xff, 0xff, 0xc0, 0x3f, 0xff, 0xc0, 0x03, 0xff, 0xff, 0xfc, 0x00, 0x7f, 0xff, 0x00, 0x00, 0x00, 0x00, 0x07, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x0e, 0x00, 0xff, 0x00, 0xff};
    byte p256Scalar[p256SMResultLength] = { 0x2a, 0x26, 0x5f, 0x8b, 0xcb, 0xdc, 0xaf, 0x94, 0xd5, 0x85, 0x19, 0x14, 0x1e, 0x57, 0x81, 0x24, 0xcb, 0x40, 0xd6, 0x4a, 0x50, 0x1f, 0xba, 0x9c, 0x11, 0x84, 0x7b, 0x28, 0x96, 0x5b, 0xc7, 0x37};
    byte p384Scalar[p384SMResultLength] = { 0x2a, 0x26, 0x5f, 0x8b, 0xcb, 0xdc, 0xaf, 0x94, 0xd5, 0x85, 0x19, 0x14, 0x1e, 0x57, 0x81, 0x24, 0xcb, 0x40, 0xd6, 0x4a, 0x50, 0x1f, 0xba, 0x9c, 0x11, 0x84, 0x7b, 0x28, 0x96, 0x5b, 0xc7, 0x37, 0x7f, 0xff, 0xff, 0xc0, 0x3f, 0xff, 0xc0, 0x03, 0xff, 0xff, 0xfc, 0x00, 0x7f, 0xff, 0x00, 0x00};
    byte p521Scalar[p521SMResultLength] = { 0x2a, 0x26, 0x5f, 0x8b, 0xcb, 0xdc, 0xaf, 0x94, 0xd5, 0x85, 0x19, 0x14, 0x1e, 0x57, 0x81, 0x24, 0xcb, 0x40, 0xd6, 0x4a, 0x50, 0x1f, 0xba, 0x9c, 0x11, 0x84, 0x7b, 0x28, 0x96, 0x5b, 0xc7, 0x37, 0x7f, 0xff, 0xff, 0xc0, 0x3f, 0xff, 0xc0, 0x03, 0xff, 0xff, 0xfc, 0x00, 0x7f, 0xff, 0x00, 0x00, 0x7f, 0xff, 0xff, 0xc0, 0x3f, 0xff, 0xc0, 0x03, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff};

    int i;
    for (i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p224XBasePointHandle,p224YBasePointHandle,p224Handle,p224XBasePointHandle,p224YBasePointHandle);
        scalarMultEC(fieldOrderHandle,basePointOrderHandle,p224Handle,p224XBasePointHandle,p224YBasePointHandle,p224Scalar,sizeof(p224Scalar));
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p256XBasePointHandle,p256YBasePointHandle,p256Handle,p256XBasePointHandle,p256YBasePointHandle);
        scalarMultEC(fieldOrderHandle,basePointOrderHandle,p256Handle,p256XBasePointHandle,p256YBasePointHandle,p256Scalar,sizeof(p256Scalar));
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p384XBasePointHandle,p384YBasePointHandle,p384Handle,p384XBasePointHandle,p384YBasePointHandle);
        scalarMultEC(fieldOrderHandle,basePointOrderHandle,p384Handle,p384XBasePointHandle,p384YBasePointHandle,p384Scalar,sizeof(p384Scalar));
    }
    for ( i = 0; i < LIMIT/4; i++)
    {
        doubleEC(p521XBasePointHandle,p521YBasePointHandle,p521Handle,p521XBasePointHandle,p521YBasePointHandle);
        scalarMultEC(fieldOrderHandle,basePointOrderHandle,p521Handle,p521XBasePointHandle,p521YBasePointHandle,p521Scalar,sizeof(p521Scalar));
    }
}
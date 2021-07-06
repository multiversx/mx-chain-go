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
const int p521SMResultLength = 128;
typedef byte P224MCRESULT[29];
typedef byte P256MCRESULT[33];
typedef byte P384MCRESULT[49];
typedef byte P521MCRESULT[67];

byte scalar[] = {
        0xff, 0x2a, 0x26, 0x5f, 0x8b, 0xcb, 0xdc, 0xaf, 
        0xd5, 0x85, 0x19, 0x14, 0x1e, 0x57, 0x81, 0x24, 
        0xcb, 0x40, 0xd6, 0x4a, 0x50, 0x1f, 0xba, 0x9c, 
        0x11, 0x84, 0x7b, 0x28, 0x96, 0x5b, 0xc7, 0x37, 
        0x7f, 0xff, 0xff, 0xc0, 0x3f, 0xff, 0xc0, 0x03, 
        0xff, 0xff, 0xfc, 0x00, 0x7f, 0xff, 0x00, 0x00, 
        0x7f, 0xff, 0xff, 0xc0, 0x3f, 0xff, 0xc0, 0x03, 
        0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 
        0xff, 0x2a, 0x26, 0x5f, 0x8b, 0xcb, 0xdc, 0xaf, 
        0xd5, 0x85, 0x19, 0x14, 0x1e, 0x57, 0x81, 0x24, 
        0xcb, 0x40, 0xd6, 0x4a, 0x50, 0x1f, 0xba, 0x9c, 
        0x11, 0x84, 0x7b, 0x28, 0x96, 0x5b, 0xc7, 0x37, 
        0x7f, 0xff, 0xff, 0xc0, 0x3f, 0xff, 0xc0, 0x03, 
        0xff, 0xff, 0xfc, 0x00, 0x7f, 0xff, 0x00, 0x00, 
        0x7f, 0xff, 0xff, 0xc0, 0x3f, 0xff, 0xc0, 0x03, 
        0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff
    };

const int numberOfReps = 10000;
const lengthOfScalar = 60;

void repeatAddEc(int numberOfTimes, int curveHandle, int xPointHandle, int yPointHandle);
void repeatDoubleEc(int numberOfTimes, int curveHandle, int xPointHandle, int yPointHandle);
void repeatDoubleAndIsOnCurve(int numberOfTimes, int curveHandle, int xPointHandle, int yPointHandle);
void repeatDoubleAndMarshal(int numberOfTimes, int curveHandle, int xPointHandle, int yPointHandle, byte marshalResult[]);
void repeatDoubleAndMarshalAndUnmarshal(int numberOfTimes, int curveHandle, int xPointHandle, int yPointHandle, byte marshalResult[], int sizeOfMarhsalResult);
void repeatDoubleAndMarshalCompressed(int numberOfTimes, int curveHandle, int xPointHandle, int yPointHandle, byte marshalResult[]);
void repeatDoubleAndMarshalAndUnmarshalCompressed(int numberOfTimes, int curveHandle, int xPointHandle, int yPointHandle, byte marshalCompressedResult[], int sizeOfMarhsalCompressedResult);
void repeatGenerateKey(int numberOfTimes, int curveHandle, int xPubKeyHandle, int yPubKeyHandle, byte privKey[]);
void repeatDoubleAndScalarMult(int numberOfTimes, int curveHandle, int xPointHandle, int yPointHandle);

void init() 
{
}

void addEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p224XBasePointHandle, p224YBasePointHandle);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p256XBasePointHandle, p256YBasePointHandle);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p384XBasePointHandle, p384YBasePointHandle);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p521XBasePointHandle, p521YBasePointHandle);
    repeatAddEc(numberOfReps/4, p224Handle, p224XBasePointHandle, p224YBasePointHandle);
    repeatAddEc(numberOfReps/4, p256Handle, p256XBasePointHandle, p256YBasePointHandle);
    repeatAddEc(numberOfReps/4, p384Handle, p384XBasePointHandle, p384YBasePointHandle);
    repeatAddEc(numberOfReps/4, p521Handle, p521XBasePointHandle, p521YBasePointHandle);
}

void p224AddEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p224XBasePointHandle, p224YBasePointHandle);
    repeatAddEc(numberOfReps, p224Handle, p224XBasePointHandle, p224YBasePointHandle);
}

void p256AddEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p256XBasePointHandle, p256YBasePointHandle);
    repeatAddEc(numberOfReps, p256Handle, p256XBasePointHandle, p256YBasePointHandle);
}

void p384AddEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p384XBasePointHandle, p384YBasePointHandle);
    repeatAddEc(numberOfReps, p384Handle, p384XBasePointHandle, p384YBasePointHandle);
}

void p521AddEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p521XBasePointHandle, p521YBasePointHandle);
    repeatAddEc(numberOfReps, p521Handle, p521XBasePointHandle, p521YBasePointHandle);
}

void doubleEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p224XBasePointHandle, p224YBasePointHandle);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p256XBasePointHandle, p256YBasePointHandle);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p384XBasePointHandle, p384YBasePointHandle);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p521XBasePointHandle, p521YBasePointHandle);
    repeatDoubleEc(numberOfReps/4, p224Handle, p224XBasePointHandle, p224YBasePointHandle);
    repeatDoubleEc(numberOfReps/4, p256Handle, p256XBasePointHandle, p256YBasePointHandle);
    repeatDoubleEc(numberOfReps/4, p384Handle, p384XBasePointHandle, p384YBasePointHandle);
    repeatDoubleEc(numberOfReps/4, p521Handle, p521XBasePointHandle, p521YBasePointHandle);
}

void p224DoubleEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p224XBasePointHandle, p224YBasePointHandle);
    repeatDoubleEc(numberOfReps, p224Handle, p224XBasePointHandle, p224YBasePointHandle);
}

void p256DoubleEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p256XBasePointHandle, p256YBasePointHandle);
    repeatDoubleEc(numberOfReps, p256Handle, p256XBasePointHandle, p256YBasePointHandle);
}

void p384DoubleEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p384XBasePointHandle, p384YBasePointHandle);
    repeatDoubleEc(numberOfReps, p384Handle, p384XBasePointHandle, p384YBasePointHandle);
}

void p521DoubleEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p521XBasePointHandle, p521YBasePointHandle);
    repeatDoubleEc(numberOfReps, p521Handle, p521XBasePointHandle, p521YBasePointHandle);
}

void isOnCurveEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p224XBasePointHandle, p224YBasePointHandle);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p256XBasePointHandle, p256YBasePointHandle);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p384XBasePointHandle, p384YBasePointHandle);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p521XBasePointHandle, p521YBasePointHandle);
    repeatDoubleAndIsOnCurve(numberOfReps/4, p224Handle, p224XBasePointHandle, p224YBasePointHandle);
    repeatDoubleAndIsOnCurve(numberOfReps/4, p256Handle, p256XBasePointHandle, p256YBasePointHandle);
    repeatDoubleAndIsOnCurve(numberOfReps/4, p384Handle, p384XBasePointHandle, p384YBasePointHandle);
    repeatDoubleAndIsOnCurve(numberOfReps/4, p521Handle, p521XBasePointHandle, p521YBasePointHandle);
}

void p224IsOnCurveEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p224XBasePointHandle, p224YBasePointHandle);
    repeatDoubleAndIsOnCurve(numberOfReps, p224Handle, p224XBasePointHandle, p224YBasePointHandle);
}

void p256IsOnCurveEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p256XBasePointHandle, p256YBasePointHandle);
    repeatDoubleAndIsOnCurve(numberOfReps, p256Handle, p256XBasePointHandle, p256YBasePointHandle);
}

void p384IsOnCurveEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p384XBasePointHandle, p384YBasePointHandle);
    repeatDoubleAndIsOnCurve(numberOfReps, p384Handle, p384XBasePointHandle, p384YBasePointHandle);
}

void p521IsOnCurveEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p521XBasePointHandle, p521YBasePointHandle);
    repeatDoubleAndIsOnCurve(numberOfReps, p521Handle, p521XBasePointHandle, p521YBasePointHandle);
}

void marshalEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p224XBasePointHandle, p224YBasePointHandle);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p256XBasePointHandle, p256YBasePointHandle);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p384XBasePointHandle, p384YBasePointHandle);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p521XBasePointHandle, p521YBasePointHandle);
    P224MRESULT p224MarshalResult;
    P256MRESULT p256MarshalResult;
    P384MRESULT p384MarshalResult;
    P521MRESULT p521MarshalResult;
    repeatDoubleAndMarshal(numberOfReps/4, p224Handle, p224XBasePointHandle, p224YBasePointHandle, p224MarshalResult);
    repeatDoubleAndMarshal(numberOfReps/4, p256Handle, p256XBasePointHandle, p256YBasePointHandle, p256MarshalResult);
    repeatDoubleAndMarshal(numberOfReps/4, p384Handle, p384XBasePointHandle, p384YBasePointHandle, p384MarshalResult);
    repeatDoubleAndMarshal(numberOfReps/4, p521Handle, p521XBasePointHandle, p521YBasePointHandle, p521MarshalResult);
}

void p224MarshalEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p224XBasePointHandle, p224YBasePointHandle);
    P224MRESULT p224MarshalResult;
    repeatDoubleAndMarshal(numberOfReps, p224Handle, p224XBasePointHandle, p224YBasePointHandle, p224MarshalResult);
}

void p256MarshalEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p256XBasePointHandle, p256YBasePointHandle);
    P256MRESULT p256MarshalResult;
    repeatDoubleAndMarshal(numberOfReps, p256Handle, p256XBasePointHandle, p256YBasePointHandle, p256MarshalResult);
}

void p384MarshalEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p384XBasePointHandle, p384YBasePointHandle);
    P384MRESULT p384MarshalResult;
    repeatDoubleAndMarshal(numberOfReps, p384Handle, p384XBasePointHandle, p384YBasePointHandle, p384MarshalResult);
}

void p521MarshalEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p521XBasePointHandle, p521YBasePointHandle);
    P521MRESULT p521MarshalResult;
    repeatDoubleAndMarshal(numberOfReps, p521Handle, p521XBasePointHandle, p521YBasePointHandle, p521MarshalResult);
}

void unmarshalEcTest() 
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p224XBasePointHandle, p224YBasePointHandle);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p256XBasePointHandle, p256YBasePointHandle);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p384XBasePointHandle, p384YBasePointHandle);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p521XBasePointHandle, p521YBasePointHandle);
    P224MRESULT p224MarshalResult;
    P256MRESULT p256MarshalResult;
    P384MRESULT p384MarshalResult;
    P521MRESULT p521MarshalResult;
    repeatDoubleAndMarshalAndUnmarshal(numberOfReps/4, p224Handle, p224XBasePointHandle, p224YBasePointHandle, p224MarshalResult, sizeof(p224MarshalResult));
    repeatDoubleAndMarshalAndUnmarshal(numberOfReps/4, p256Handle, p256XBasePointHandle, p256YBasePointHandle, p256MarshalResult, sizeof(p256MarshalResult));
    repeatDoubleAndMarshalAndUnmarshal(numberOfReps/4, p384Handle, p384XBasePointHandle, p384YBasePointHandle, p384MarshalResult, sizeof(p384MarshalResult));
    repeatDoubleAndMarshalAndUnmarshal(numberOfReps/4, p521Handle, p521XBasePointHandle, p521YBasePointHandle, p521MarshalResult, sizeof(p521MarshalResult));
}

void p224UnmarshalEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p224XBasePointHandle, p224YBasePointHandle);
    P224MRESULT p224MarshalResult;
    repeatDoubleAndMarshalAndUnmarshal(numberOfReps, p224Handle, p224XBasePointHandle, p224YBasePointHandle, p224MarshalResult, sizeof(p224MarshalResult));
}

void p256UnmarshalEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p256XBasePointHandle, p256YBasePointHandle);
    P256MRESULT p256MarshalResult;
    repeatDoubleAndMarshalAndUnmarshal(numberOfReps, p256Handle, p256XBasePointHandle, p256YBasePointHandle, p256MarshalResult, sizeof(p256MarshalResult));
}

void p384UnmarshalEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p384XBasePointHandle, p384YBasePointHandle);
    P384MRESULT p384MarshalResult;
    repeatDoubleAndMarshalAndUnmarshal(numberOfReps, p384Handle, p384XBasePointHandle, p384YBasePointHandle, p384MarshalResult, sizeof(p384MarshalResult));
}

void p521UnmarshalEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p521XBasePointHandle, p521YBasePointHandle);
    P521MRESULT p521MarshalResult;
    repeatDoubleAndMarshalAndUnmarshal(numberOfReps, p521Handle, p521XBasePointHandle, p521YBasePointHandle, p521MarshalResult, sizeof(p521MarshalResult));
}

void marshalCompressedEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p224XBasePointHandle, p224YBasePointHandle);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p256XBasePointHandle, p256YBasePointHandle);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p384XBasePointHandle, p384YBasePointHandle);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p521XBasePointHandle, p521YBasePointHandle);
    P224MCRESULT p224MarshalCompressedResult;
    P256MCRESULT p256MarshalCompressedResult;
    P384MCRESULT p384MarshalCompressedResult;
    P521MCRESULT p521MarshalCompressedResult;
    repeatDoubleAndMarshalCompressed(numberOfReps/4, p224Handle, p224XBasePointHandle, p224YBasePointHandle, p224MarshalCompressedResult);
    repeatDoubleAndMarshalCompressed(numberOfReps/4, p256Handle, p256XBasePointHandle, p256YBasePointHandle, p256MarshalCompressedResult);
    repeatDoubleAndMarshalCompressed(numberOfReps/4, p384Handle, p384XBasePointHandle, p384YBasePointHandle, p384MarshalCompressedResult);
    repeatDoubleAndMarshalCompressed(numberOfReps/4, p521Handle, p521XBasePointHandle, p521YBasePointHandle, p521MarshalCompressedResult);
}

void p224MarshalCompressedEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p224XBasePointHandle, p224YBasePointHandle);
    P224MCRESULT p224MarshalCompressedResult;
    repeatDoubleAndMarshalCompressed(numberOfReps, p224Handle, p224XBasePointHandle, p224YBasePointHandle, p224MarshalCompressedResult);
}

void p256MarshalCompressedEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p256XBasePointHandle, p256YBasePointHandle);
    P256MCRESULT p256MarshalCompressedResult;
    repeatDoubleAndMarshalCompressed(numberOfReps, p256Handle, p256XBasePointHandle, p256YBasePointHandle, p256MarshalCompressedResult);
}

void p384MarshalCompressedEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p384XBasePointHandle, p384YBasePointHandle);
    P384MCRESULT p384MarshalCompressedResult;
    repeatDoubleAndMarshalCompressed(numberOfReps, p384Handle, p384XBasePointHandle, p384YBasePointHandle, p384MarshalCompressedResult);
}

void p521MarshalCompressedEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p521XBasePointHandle, p521YBasePointHandle);
    P521MCRESULT p521MarshalCompressedResult;
    repeatDoubleAndMarshalCompressed(numberOfReps, p521Handle, p521XBasePointHandle, p521YBasePointHandle, p521MarshalCompressedResult);
}

void unmarshalCompressedEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p224XBasePointHandle, p224YBasePointHandle);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p256XBasePointHandle, p256YBasePointHandle);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p384XBasePointHandle, p384YBasePointHandle);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p521XBasePointHandle, p521YBasePointHandle);
    P224MCRESULT p224MarshalCompressedResult;
    P256MCRESULT p256MarshalCompressedResult;
    P384MCRESULT p384MarshalCompressedResult;
    P521MCRESULT p521MarshalCompressedResult;
    repeatDoubleAndMarshalAndUnmarshalCompressed(numberOfReps/4, p224Handle, p224XBasePointHandle, p224YBasePointHandle, p224MarshalCompressedResult, sizeof(p224MarshalCompressedResult));
    repeatDoubleAndMarshalAndUnmarshalCompressed(numberOfReps/4, p256Handle, p256XBasePointHandle, p256YBasePointHandle, p256MarshalCompressedResult, sizeof(p256MarshalCompressedResult));
    repeatDoubleAndMarshalAndUnmarshalCompressed(numberOfReps/4, p384Handle, p384XBasePointHandle, p384YBasePointHandle, p384MarshalCompressedResult, sizeof(p384MarshalCompressedResult));
    repeatDoubleAndMarshalAndUnmarshalCompressed(numberOfReps/4, p521Handle, p521XBasePointHandle, p521YBasePointHandle, p521MarshalCompressedResult, sizeof(p521MarshalCompressedResult));
}

void p224UnmarshalCompressedEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p224XBasePointHandle, p224YBasePointHandle);
    P224MCRESULT p224MarshalCompressedResult;
    repeatDoubleAndMarshalAndUnmarshalCompressed(numberOfReps, p224Handle, p224XBasePointHandle, p224YBasePointHandle, p224MarshalCompressedResult, sizeof(p224MarshalCompressedResult));
}

void p256UnmarshalCompressedEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p256XBasePointHandle, p256YBasePointHandle);
    P256MCRESULT p256MarshalCompressedResult;
    repeatDoubleAndMarshalAndUnmarshalCompressed(numberOfReps, p256Handle, p256XBasePointHandle, p256YBasePointHandle, p256MarshalCompressedResult, sizeof(p256MarshalCompressedResult));
}

void p384UnmarshalCompressedEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p384XBasePointHandle, p384YBasePointHandle);
    P384MCRESULT p384MarshalCompressedResult;
    repeatDoubleAndMarshalAndUnmarshalCompressed(numberOfReps, p384Handle, p384XBasePointHandle, p384YBasePointHandle, p384MarshalCompressedResult, sizeof(p384MarshalCompressedResult));
}

void p521UnmarshalCompressedEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p521XBasePointHandle, p521YBasePointHandle);
    P521MCRESULT p521MarshalCompressedResult;
    repeatDoubleAndMarshalAndUnmarshalCompressed(numberOfReps, p521Handle, p521XBasePointHandle, p521YBasePointHandle, p521MarshalCompressedResult, sizeof(p521MarshalCompressedResult));
}

void generateKeyEcTest()
{
    int xPubKeyHandle = bigIntNew(0);
    int yPubKeyHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p256Handle = p256Ec();
    int p384Handle = p384Ec();
    int p521Handle = p521Ec();
    
    P224PRIVKEY p224PrivKey;
    P256PRIVKEY p256PrivKey;
    P384PRIVKEY p384PrivKey;
    P521PRIVKEY p521PrivKey;
    repeatGenerateKey(numberOfReps/4, p224Handle, xPubKeyHandle, yPubKeyHandle, p224PrivKey);
    repeatGenerateKey(numberOfReps/4, p256Handle, xPubKeyHandle, yPubKeyHandle, p256PrivKey);
    repeatGenerateKey(numberOfReps/4, p384Handle, xPubKeyHandle, yPubKeyHandle, p384PrivKey);
    repeatGenerateKey(numberOfReps/4, p521Handle, xPubKeyHandle, yPubKeyHandle, p521PrivKey);
}

void p224GenerateKeyEcTest()
{
    int xPubKeyHandle = bigIntNew(0);
    int yPubKeyHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    P224PRIVKEY p224PrivKey;
    repeatGenerateKey(numberOfReps, p224Handle, xPubKeyHandle, yPubKeyHandle, p224PrivKey);
}

void p256GenerateKeyEcTest()
{
    int xPubKeyHandle = bigIntNew(0);
    int yPubKeyHandle = bigIntNew(0);
    int p256Handle = p256Ec();
    P256PRIVKEY p256PrivKey;
    repeatGenerateKey(numberOfReps, p256Handle, xPubKeyHandle, yPubKeyHandle, p256PrivKey);
}

void p384GenerateKeyEcTest()
{
    int xPubKeyHandle = bigIntNew(0);
    int yPubKeyHandle = bigIntNew(0);
    int p384Handle = p384Ec();
    P384PRIVKEY p384PrivKey;
    repeatGenerateKey(numberOfReps, p384Handle, xPubKeyHandle, yPubKeyHandle, p384PrivKey);
}

void p521GenerateKeyEcTest()
{
    int xPubKeyHandle = bigIntNew(0);
    int yPubKeyHandle = bigIntNew(0);
    int p521Handle = p521Ec();
    P521PRIVKEY p521PrivKey;
    repeatGenerateKey(numberOfReps, p521Handle, xPubKeyHandle, yPubKeyHandle, p521PrivKey);
}

void scalarMultEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p224XBasePointHandle, p224YBasePointHandle);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p256XBasePointHandle, p256YBasePointHandle);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p384XBasePointHandle, p384YBasePointHandle);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p521XBasePointHandle, p521YBasePointHandle);
    repeatDoubleAndScalarMult(numberOfReps/4, p224Handle, p224XBasePointHandle, p224YBasePointHandle);
    repeatDoubleAndScalarMult(numberOfReps/4, p256Handle, p256XBasePointHandle, p256YBasePointHandle);
    repeatDoubleAndScalarMult(numberOfReps/4, p384Handle, p384XBasePointHandle, p384YBasePointHandle);
    repeatDoubleAndScalarMult(numberOfReps/4, p521Handle, p521XBasePointHandle, p521YBasePointHandle);
}

void p224ScalarMultEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p224Handle = p224Ec();
    int p224XBasePointHandle = bigIntNew(0);
    int p224YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p224Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p224XBasePointHandle, p224YBasePointHandle);
    repeatDoubleAndScalarMult(numberOfReps, p224Handle, p224XBasePointHandle, p224YBasePointHandle);
}

void p256ScalarMultEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p256Handle = p256Ec();
    int p256XBasePointHandle = bigIntNew(0);
    int p256YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p256Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p256XBasePointHandle, p256YBasePointHandle);
    repeatDoubleAndScalarMult(numberOfReps, p256Handle, p256XBasePointHandle, p256YBasePointHandle);
}

void p384ScalarMultEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p384Handle = p384Ec();
    int p384XBasePointHandle = bigIntNew(0);
    int p384YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p384Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p384XBasePointHandle, p384YBasePointHandle);
    repeatDoubleAndScalarMult(numberOfReps, p384Handle, p384XBasePointHandle, p384YBasePointHandle);
}

void p521ScalarMultEcTest()
{
    int fieldOrderHandle = bigIntNew(0);
    int basePointOrderHandle = bigIntNew(0);
    int eqConstantHandle = bigIntNew(0);
    int p521Handle = p521Ec();
    int p521XBasePointHandle = bigIntNew(0);
    int p521YBasePointHandle = bigIntNew(0);
    ellipticCurveGetValues(p521Handle, fieldOrderHandle, basePointOrderHandle, eqConstantHandle, p521XBasePointHandle, p521YBasePointHandle);
    repeatDoubleAndScalarMult(numberOfReps, p521Handle, p521XBasePointHandle, p521YBasePointHandle);
}

void repeatAddEc(int numberOfTimes, int curveHandle, int xPointHandle, int yPointHandle)
{
    for(int i = 0; i < numberOfTimes; i++)
    {
        addEC(xPointHandle, yPointHandle, curveHandle, xPointHandle, yPointHandle, xPointHandle, yPointHandle);
    }
}

void repeatDoubleEc(int numberOfTimes, int curveHandle, int xPointHandle, int yPointHandle)
{
    for (int i = 0; i < numberOfTimes; i++)
    {
        doubleEC(xPointHandle, yPointHandle, curveHandle, xPointHandle, yPointHandle);
    }
}

void repeatDoubleAndIsOnCurve(int numberOfTimes, int curveHandle, int xPointHandle, int yPointHandle)
{
    for (int i = 0; i < numberOfTimes; i++)
    {
        doubleEC(xPointHandle, yPointHandle, curveHandle, xPointHandle, yPointHandle);
        isOnCurveEC(curveHandle, xPointHandle, yPointHandle);
    }
}

void repeatDoubleAndMarshal(int numberOfTimes, int curveHandle, int xPointHandle, int yPointHandle, byte marshalResult[])
{
    for (int i = 0; i < numberOfTimes; i++)
    {
        doubleEC(xPointHandle, yPointHandle, curveHandle, xPointHandle, yPointHandle);
        marshalEC(xPointHandle, yPointHandle, curveHandle, marshalResult);
    }
}

void repeatDoubleAndMarshalCompressed(int numberOfTimes, int curveHandle, int xPointHandle, int yPointHandle, byte marshalResult[])
{
    for (int i = 0; i < numberOfTimes; i++)
    {
        doubleEC(xPointHandle, yPointHandle, curveHandle, xPointHandle, yPointHandle);
        marshalCompressedEC(xPointHandle, yPointHandle, curveHandle, marshalResult);
    }
}

void repeatDoubleAndMarshalAndUnmarshal(int numberOfTimes, int curveHandle, int xPointHandle, int yPointHandle, byte marshalResult[], int sizeOfMarhsalResult)
{
    for (int i = 0; i < numberOfTimes; i++)
    {
        doubleEC(xPointHandle, yPointHandle, curveHandle, xPointHandle, yPointHandle);
        marshalEC(xPointHandle, yPointHandle, curveHandle, marshalResult);
        unmarshalEC(xPointHandle, yPointHandle, curveHandle, marshalResult, sizeOfMarhsalResult);
    }
}

void repeatDoubleAndMarshalAndUnmarshalCompressed(int numberOfTimes, int curveHandle, int xPointHandle, int yPointHandle, byte marshalCompressedResult[], int sizeOfMarhsalCompressedResult)
{
    for (int i = 0; i < numberOfTimes; i++)
    {
        doubleEC(xPointHandle, yPointHandle, curveHandle, xPointHandle, yPointHandle);
        marshalCompressedEC(xPointHandle, yPointHandle, curveHandle, marshalCompressedResult);
        unmarshalCompressedEC(xPointHandle, yPointHandle, curveHandle, marshalCompressedResult, sizeOfMarhsalCompressedResult);
    }
}

void repeatGenerateKey(int numberOfTimes, int curveHandle, int xPubKeyHandle, int yPubKeyHandle, byte privKey[])
{
    for (int i = 0; i < numberOfTimes; i++)
    {
        generateKeyEC(xPubKeyHandle, yPubKeyHandle, curveHandle, privKey);
    }
}

void repeatDoubleAndScalarMult(int numberOfTimes, int curveHandle, int xPointHandle, int yPointHandle)
{
    int xResult = bigIntNew(0);
    int yResult = bigIntNew(0);
    for (int i = 0; i < numberOfTimes; i++)
    {
        doubleEC(xPointHandle, yPointHandle, curveHandle, xPointHandle, yPointHandle);
        scalarMultEC(xResult, yResult, curveHandle, xPointHandle, yPointHandle, scalar, lengthOfScalar);
    }
}
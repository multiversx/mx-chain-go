#pragma once
/**
	@file
	@brief C interface of bls.hpp
	@author MITSUNARI Shigeo(@herumi)
	@license modified new BSD license
	http://opensource.org/licenses/BSD-3-Clause
*/
#define MCLBN_NO_AUTOLINK
#include <mcl/bn.h>

#ifdef BLS_ETH
	#ifndef BLS_SWAP_G
		#define BLS_SWAP_G
	#endif
	#define BLS_COMPILER_TIME_VAR_ADJ 200
#endif
#ifdef BLS_SWAP_G
	#ifndef BLS_COMPILER_TIME_VAR_ADJ
		#define BLS_COMPILER_TIME_VAR_ADJ 100
	#endif
	/*
		error if BLS_SWAP_G is inconsistently used between library and exe
	*/
	#undef MCLBN_COMPILED_TIME_VAR
	#define MCLBN_COMPILED_TIME_VAR ((MCLBN_FR_UNIT_SIZE) * 10 + (MCLBN_FP_UNIT_SIZE) + BLS_COMPILER_TIME_VAR_ADJ)
#endif

#ifdef _MSC_VER
	#ifdef BLS_DONT_EXPORT
		#define BLS_DLL_API
	#else
		#ifdef BLS_DLL_EXPORT
			#define BLS_DLL_API __declspec(dllexport)
		#else
			#define BLS_DLL_API __declspec(dllimport)
		#endif
	#endif
	#ifndef BLS_NO_AUTOLINK
		#if MCLBN_FP_UNIT_SIZE == 4
			#pragma comment(lib, "bls256.lib")
		#elif (MCLBN_FP_UNIT_SIZE == 6) && (MCLBN_FR_UNIT_SIZE == 4)
			#pragma comment(lib, "bls384_256.lib")
		#elif (MCLBN_FP_UNIT_SIZE == 6) && (MCLBN_FR_UNIT_SIZE == 6)
			#pragma comment(lib, "bls384.lib")
		#endif
	#endif
#elif defined(__EMSCRIPTEN__) && !defined(BLS_DONT_EXPORT)
	#define BLS_DLL_API __attribute__((used))
#elif defined(__wasm__) && !defined(BLS_DONT_EXPORT)
	#define BLS_DLL_API __attribute__((visibility("default")))
#else
	#define BLS_DLL_API
#endif

#ifdef __cplusplus
extern "C" {
#endif

typedef struct {
	mclBnFr v;
} blsId;

typedef struct {
	mclBnFr v;
} blsSecretKey;

typedef struct {
#ifdef BLS_SWAP_G
	mclBnG1 v;
#else
	mclBnG2 v;
#endif
} blsPublicKey;

typedef struct {
#ifdef BLS_SWAP_G
	mclBnG2 v;
#else
	mclBnG1 v;
#endif
} blsSignature;

/*
	initialize this library
	call this once before using the other functions
	@param curve [in] enum value defined in mcl/bn.h
	@param compiledTimeVar [in] specify MCLBN_COMPILED_TIME_VAR,
	which macro is used to make sure that the values
	are the same when the library is built and used
	@return 0 if success
	@note blsInit() is not thread safe
*/
BLS_DLL_API int blsInit(int curve, int compiledTimeVar);

/*
	use new eth 2.0 spec
	@return 0 if success
	@remark
	this functions and the spec may change until it is fixed
	the size of message <= 32
*/
#define BLS_ETH_MODE_OLD 0
#define BLS_ETH_MODE_LATEST 1
BLS_DLL_API int blsSetETHmode(int mode);

/*
	set ETH serialization mode for BLS12-381
	@param ETHserialization [in] 1:enable,  0:disable
	@note ignore the flag if curve is not BLS12-381
	@note set in blsInit if BLS_ETH is defined
*/
BLS_DLL_API void blsSetETHserialization(int ETHserialization);

BLS_DLL_API void blsIdSetInt(blsId *id, int x);

// sec = buf & (1 << bitLen(r)) - 1
// if (sec >= r) sec &= (1 << (bitLen(r) - 1)) - 1
// always return 0
BLS_DLL_API int blsSecretKeySetLittleEndian(blsSecretKey *sec, const void *buf, mclSize bufSize);
// return 0 if success (bufSize <= 64) else -1
// set (buf mod r) to sec
BLS_DLL_API int blsSecretKeySetLittleEndianMod(blsSecretKey *sec, const void *buf, mclSize bufSize);

BLS_DLL_API void blsGetPublicKey(blsPublicKey *pub, const blsSecretKey *sec);

// calculate the has of m and sign the hash
BLS_DLL_API void blsSign(blsSignature *sig, const blsSecretKey *sec, const void *m, mclSize size);

BLS_DLL_API int blsPublicKeyToG1(const blsPublicKey *pub, mclBnG1 *x);
BLS_DLL_API int blsG1ToPublicKey(const mclBnG1 *x, blsPublicKey *pub);
BLS_DLL_API int blsSignatureToG1(const blsSignature *sig, mclBnG1 *x);
BLS_DLL_API int blsG1ToSignature(const mclBnG1 *x, blsSignature *sig);
BLS_DLL_API int blsPublicKeyToG2(const blsPublicKey *pub, mclBnG2 *x);
BLS_DLL_API int blsG2ToPublicKey(const mclBnG2 *x, blsPublicKey *pub);
BLS_DLL_API int blsSignatureToG2(const blsSignature *sig, mclBnG2 *x);
BLS_DLL_API int blsG2ToSignature(const mclBnG2 *x, blsSignature *sig);
BLS_DLL_API void blsSecretKeyToFr(const blsSecretKey *sec, mclBnFr *s);
BLS_DLL_API void blsFrToSecretKey(const mclBnFr *s, blsSecretKey *sec);

// return 1 if valid
BLS_DLL_API int blsVerify(const blsSignature *sig, const blsPublicKey *pub, const void *m, mclSize size);

// aggSig = sum of sigVec[0..n]
BLS_DLL_API void blsAggregateSignature(blsSignature *aggSig, const blsSignature *sigVec, mclSize n);

// verify(sig, sum of pubVec[0..n], msg)
BLS_DLL_API int blsFastAggregateVerify(const blsSignature *sig, const blsPublicKey *pubVec, mclSize n, const void *msg, mclSize msgSize);

/*
	all msg[i] has the same msgSize byte, so msgVec must have (msgSize * n) byte area
	verify prod e(H(pubVec[i], msgToG2[i]) == e(P, sig)
	@note CHECK that sig has the valid order, all msg are different each other before calling this
*/
BLS_DLL_API int blsAggregateVerifyNoCheck(const blsSignature *sig, const blsPublicKey *pubVec, const void *msgVec, mclSize msgSize, mclSize n);

// return written byte size if success else 0
BLS_DLL_API mclSize blsIdSerialize(void *buf, mclSize maxBufSize, const blsId *id);
BLS_DLL_API mclSize blsSecretKeySerialize(void *buf, mclSize maxBufSize, const blsSecretKey *sec);
BLS_DLL_API mclSize blsPublicKeySerialize(void *buf, mclSize maxBufSize, const blsPublicKey *pub);
BLS_DLL_API mclSize blsSignatureSerialize(void *buf, mclSize maxBufSize, const blsSignature *sig);

// return read byte size if success else 0
BLS_DLL_API mclSize blsIdDeserialize(blsId *id, const void *buf, mclSize bufSize);
BLS_DLL_API mclSize blsSecretKeyDeserialize(blsSecretKey *sec, const void *buf, mclSize bufSize);
BLS_DLL_API mclSize blsPublicKeyDeserialize(blsPublicKey *pub, const void *buf, mclSize bufSize);
BLS_DLL_API mclSize blsSignatureDeserialize(blsSignature *sig, const void *buf, mclSize bufSize);

// return 1 if same else 0
BLS_DLL_API int blsIdIsEqual(const blsId *lhs, const blsId *rhs);
BLS_DLL_API int blsSecretKeyIsEqual(const blsSecretKey *lhs, const blsSecretKey *rhs);
BLS_DLL_API int blsPublicKeyIsEqual(const blsPublicKey *lhs, const blsPublicKey *rhs);
BLS_DLL_API int blsSignatureIsEqual(const blsSignature *lhs, const blsSignature *rhs);

// return 0 if success
BLS_DLL_API int blsSecretKeyShare(blsSecretKey *sec, const blsSecretKey* msk, mclSize k, const blsId *id);
BLS_DLL_API int blsPublicKeyShare(blsPublicKey *pub, const blsPublicKey *mpk, mclSize k, const blsId *id);

BLS_DLL_API int blsSecretKeyRecover(blsSecretKey *sec, const blsSecretKey *secVec, const blsId *idVec, mclSize n);
BLS_DLL_API int blsPublicKeyRecover(blsPublicKey *pub, const blsPublicKey *pubVec, const blsId *idVec, mclSize n);
BLS_DLL_API int blsSignatureRecover(blsSignature *sig, const blsSignature *sigVec, const blsId *idVec, mclSize n);

// add
BLS_DLL_API void blsSecretKeyAdd(blsSecretKey *sec, const blsSecretKey *rhs);
BLS_DLL_API void blsPublicKeyAdd(blsPublicKey *pub, const blsPublicKey *rhs);
BLS_DLL_API void blsSignatureAdd(blsSignature *sig, const blsSignature *rhs);

/*
	verify whether a point of an elliptic curve has order r
	This api affetcs setStr(), deserialize() for G2 on BN or G1/G2 on BLS12
	@param doVerify [in] does not verify if zero(default 1)
	Signature = G1, PublicKey = G2
*/
BLS_DLL_API void blsSignatureVerifyOrder(int doVerify);
BLS_DLL_API void blsPublicKeyVerifyOrder(int doVerify);
//	deserialize under VerifyOrder(true) = deserialize under VerifyOrder(false) + IsValidOrder
BLS_DLL_API int blsSignatureIsValidOrder(const blsSignature *sig);
BLS_DLL_API int blsPublicKeyIsValidOrder(const blsPublicKey *pub);

#ifndef BLS_MINIMUM_API

/*
	verify X == sY by checking e(X, sQ) = e(Y, Q)
	@param X [in]
	@param Y [in]
	@param pub [in] pub = sQ
	@return 1 if e(X, pub) = e(Y, Q) else 0
*/
BLS_DLL_API int blsVerifyPairing(const blsSignature *X, const blsSignature *Y, const blsPublicKey *pub);

/*
	sign the hash
	use the low (bitSize of r) - 1 bit of h
	return 0 if success else -1
	NOTE : return false if h is zero or c1 or -c1 value for BN254. see hashTest() in test/bls_test.hpp
*/
BLS_DLL_API int blsSignHash(blsSignature *sig, const blsSecretKey *sec, const void *h, mclSize size);
// return 1 if valid
BLS_DLL_API int blsVerifyHash(const blsSignature *sig, const blsPublicKey *pub, const void *h, mclSize size);

/*
	verify aggSig with pubVec[0, n) and hVec[0, n)
	e(aggSig, Q) = prod_i e(hVec[i], pubVec[i])
	return 1 if valid
	@note do not check duplication of hVec
*/
BLS_DLL_API int blsVerifyAggregatedHashes(const blsSignature *aggSig, const blsPublicKey *pubVec, const void *hVec, size_t sizeofHash, mclSize n);

///// from here only for BLS12-381 with BLS_ETH
/*
	sign hashWithDomain by sec
	hashWithDomain[0:32] 32 bytes message
	hashWithDomain[32:40] 8 bytes data
	see https://github.com/ethereum/eth2.0-specs/blob/dev/specs/bls_signature.md#hash_to_g2
	HashWithDomain apis support only for BLS_ETH=1 and BLS12_381
	return 0 if success else -1
*/
BLS_DLL_API int blsSignHashWithDomain(blsSignature *sig, const blsSecretKey *sec, const unsigned char hashWithDomain[40]);
// return 1 if valid
BLS_DLL_API int blsVerifyHashWithDomain(const blsSignature *sig, const blsPublicKey *pub, const unsigned char hashWithDomain[40]);

/*
	Uncompressed version of Serialize/Deserialize
	the buffer size is twice of Serialize/Deserialize
*/
BLS_DLL_API mclSize blsPublicKeySerializeUncompressed(void *buf, mclSize maxBufSize, const blsPublicKey *pub);
BLS_DLL_API mclSize blsSignatureSerializeUncompressed(void *buf, mclSize maxBufSize, const blsSignature *sig);
BLS_DLL_API mclSize blsPublicKeyDeserializeUncompressed(blsPublicKey *pub, const void *buf, mclSize bufSize);
BLS_DLL_API mclSize blsSignatureDeserializeUncompressed(blsSignature *sig, const void *buf, mclSize bufSize);

/*
	pubVec is an array of size n
	hashWithDomain is an array of size (40 * n)
*/
BLS_DLL_API int blsVerifyAggregatedHashWithDomain(const blsSignature *aggSig, const blsPublicKey *pubVec, const unsigned char hashWithDomain[][40], mclSize n);

///// to here only for BLS12-381 with BLS_ETH

// sub
BLS_DLL_API void blsSecretKeySub(blsSecretKey *sec, const blsSecretKey *rhs);
BLS_DLL_API void blsPublicKeySub(blsPublicKey *pub, const blsPublicKey *rhs);
BLS_DLL_API void blsSignatureSub(blsSignature *sig, const blsSignature *rhs);

// not thread safe version (old blsInit)
BLS_DLL_API int blsInitNotThreadSafe(int curve, int compiledTimeVar);

BLS_DLL_API mclSize blsGetOpUnitSize(void);
// return strlen(buf) if success else 0
BLS_DLL_API int blsGetCurveOrder(char *buf, mclSize maxBufSize);
BLS_DLL_API int blsGetFieldOrder(char *buf, mclSize maxBufSize);

// return serialized secretKey size
BLS_DLL_API int blsGetSerializedSecretKeyByteSize(void);
// return serialized publicKey size
BLS_DLL_API int blsGetSerializedPublicKeyByteSize(void);
// return serialized signature size
BLS_DLL_API int blsGetSerializedSignatureByteSize(void);

// return bytes for serialized G1(=Fp)
BLS_DLL_API int blsGetG1ByteSize(void);

// return bytes for serialized Fr
BLS_DLL_API int blsGetFrByteSize(void);

// get a generator of PublicKey
BLS_DLL_API void blsGetGeneratorOfPublicKey(blsPublicKey *pub);

// return 0 if success
BLS_DLL_API int blsIdSetDecStr(blsId *id, const char *buf, mclSize bufSize);
BLS_DLL_API int blsIdSetHexStr(blsId *id, const char *buf, mclSize bufSize);

/*
	return strlen(buf) if success else 0
	buf is '\0' terminated
*/
BLS_DLL_API mclSize blsIdGetDecStr(char *buf, mclSize maxBufSize, const blsId *id);
BLS_DLL_API mclSize blsIdGetHexStr(char *buf, mclSize maxBufSize, const blsId *id);

// hash buf and set SecretKey
BLS_DLL_API int blsHashToSecretKey(blsSecretKey *sec, const void *buf, mclSize bufSize);
// hash buf and set Signature
BLS_DLL_API int blsHashToSignature(blsSignature *sig, const void *buf, mclSize bufSize);
#ifndef MCL_DONT_USE_CSPRNG
/*
	set secretKey if system has /dev/urandom or CryptGenRandom
	return 0 if success else -1
*/
BLS_DLL_API int blsSecretKeySetByCSPRNG(blsSecretKey *sec);
/*
	set user-defined random function for setByCSPRNG
	@param self [in] user-defined pointer
	@param readFunc [in] user-defined function,
	which writes random bufSize bytes to buf and returns bufSize if success else returns 0
	@note if self == 0 and readFunc == 0 then set default random function
	@note not threadsafe
*/
BLS_DLL_API void blsSetRandFunc(void *self, unsigned int (*readFunc)(void *self, void *buf, unsigned int bufSize));
#endif

BLS_DLL_API void blsGetPop(blsSignature *sig, const blsSecretKey *sec);

BLS_DLL_API int blsVerifyPop(const blsSignature *sig, const blsPublicKey *pub);
//////////////////////////////////////////////////////////////////////////
// the following apis will be removed

// mask buf with (1 << (bitLen(r) - 1)) - 1 if buf >= r
BLS_DLL_API int blsIdSetLittleEndian(blsId *id, const void *buf, mclSize bufSize);
/*
	return written byte size if success else 0
*/
BLS_DLL_API mclSize blsIdGetLittleEndian(void *buf, mclSize maxBufSize, const blsId *id);

// return 0 if success
BLS_DLL_API int blsSecretKeySetDecStr(blsSecretKey *sec, const char *buf, mclSize bufSize);
BLS_DLL_API int blsSecretKeySetHexStr(blsSecretKey *sec, const char *buf, mclSize bufSize);
/*
	return written byte size if success else 0
*/
BLS_DLL_API mclSize blsSecretKeyGetLittleEndian(void *buf, mclSize maxBufSize, const blsSecretKey *sec);
/*
	return strlen(buf) if success else 0
	buf is '\0' terminated
*/
BLS_DLL_API mclSize blsSecretKeyGetDecStr(char *buf, mclSize maxBufSize, const blsSecretKey *sec);
BLS_DLL_API mclSize blsSecretKeyGetHexStr(char *buf, mclSize maxBufSize, const blsSecretKey *sec);
BLS_DLL_API int blsPublicKeySetHexStr(blsPublicKey *pub, const char *buf, mclSize bufSize);
BLS_DLL_API mclSize blsPublicKeyGetHexStr(char *buf, mclSize maxBufSize, const blsPublicKey *pub);
BLS_DLL_API int blsSignatureSetHexStr(blsSignature *sig, const char *buf, mclSize bufSize);
BLS_DLL_API mclSize blsSignatureGetHexStr(char *buf, mclSize maxBufSize, const blsSignature *sig);

/*
	Diffie Hellman key exchange
	out = sec * pub
*/
BLS_DLL_API void blsDHKeyExchange(blsPublicKey *out, const blsSecretKey *sec, const blsPublicKey *pub);

#endif // BLS_MINIMUM_API

#ifdef __cplusplus
}
#endif

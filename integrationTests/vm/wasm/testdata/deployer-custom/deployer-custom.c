#include "../mxvm/context.h"
#include "../mxvm/test_utils.h"
#include "../mxvm/args.h"

byte childGeneratedAddress[32] = {};
byte childCode[5000] = {};
byte childMetadata[16] = {};

byte deploymentValue[] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42};

void init()
{
}

void deployChild()
{
	int codeSize = getArgument(0, childCode);
	int metadataSize = getArgument(1, childMetadata);

	int result = createContract(
			500000,
			deploymentValue,
			childCode,
			childMetadata,
			codeSize,
			childGeneratedAddress,
			0,
		  0,
			0);

	int isChildContract = isSmartContract(childGeneratedAddress);
	if (isChildContract == 0) {
		byte message[] = "child not a contract";
		signalError(message, sizeof(message) - 1);
	}

  finish(childGeneratedAddress, 32);
}

void upgradeChild() {
  byte childAddress[32] = {};

	int addressSize = getArgument(0, childAddress);	
  if (addressSize != 32) {
		byte message[] = "wrong or missing child address";
		signalError(message, sizeof(message) - 1);
  }

	int codeSize = getArgument(1, childCode);	
  int metadataSize = getArgument(2, childMetadata);

  upgradeContract(
      childAddress,
      10000000,
      deploymentValue,
      childCode,
      childMetadata,
      codeSize,
      0,
      0,
      0);
}

typedef unsigned char byte;
typedef unsigned int u32;
typedef int i32;

// Test finish()
void finish(byte *data, int length);
void SCMethod_FinishNil() {
}
void SCMethod_Finish42() {
	byte myNumber = 42;
	finish(&myNumber, sizeof(byte));
}

// Test getOwner() (doesn't work yet)
void getOwner(byte *address); 
byte owner_address[32] = {
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0
};
void SCMethod_FinishOwnerAddress() {
	getOwner(owner_address);
	finish(owner_address, 32);
}

// Test getCaller()
void getCaller(byte *address); 
byte caller_address[32] = {
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0
};
void SCMethod_FinishCallerAddress() {
	getCaller(caller_address);
	finish(caller_address, 32);
}

// Test getCallValue()
int getCallValue(byte* size);
byte call_value[8] = {0, 0, 0, 0, 0, 0, 0, 0};
void SCMethod_FinishCallValue() {
	int size = getCallValue(call_value);
	finish(call_value, size);
}

// Test getFunction
int getFunction(byte* name);
byte function_name[31] = "thirty.characters.............";
void SCMethod_FinishFunctionName() {
	int length = getFunction(function_name);
	function_name[29] = '-';
	finish(function_name, length + 3);
}

// Misc
//
byte SmallArray[8] = {0, 0, 0, 0, 0, 0, 0, 0};
int Count = 7;

void incrementArrayValues(byte *array, int count) {
	int counter;
	for (counter = 0; counter < count; counter++) {
		array[counter] = array[counter] + 1;
	}
}

void SCMethod_IncrementSmallArray() {
	incrementArrayValues(SmallArray, Count);
	finish(SmallArray, 8);
}


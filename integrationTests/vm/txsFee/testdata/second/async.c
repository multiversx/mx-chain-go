#include "elrond/context.h"
#include "elrond/util.h"
#include "elrond/types.h"

typedef byte ADDRESS[32];

ADDRESS zero =  {0};
ADDRESS firstScAddress = {0, 0, 0, 0, 0, 0, 0, 0, 5, 0, 93, 61, 83, 181, 208, 252, 240, 125, 34, 33, 112, 151, 137, 50, 22, 110, 233, 243, 151, 45, 48, 48};

void init() 
{
    if (getNumArguments() == 1 && getArgumentLength(0) == sizeof(ADDRESS))
    {
        getArgument(0, firstScAddress);
    }
}

void doSomething()
{
    asyncCall(firstScAddress, zero, (byte*)"callMe@01", sizeof("callMe@01") - 1);
}

void callBack() 
{

}

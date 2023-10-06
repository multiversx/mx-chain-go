#!/usr/bin/env bash

GREEN='\x1B[0;32m'
NC='\x1B[0m'

echo -e "${GREEN}Making sure any previous deployment is stopped...${NC}"
./stop.sh

echo -e "${GREEN}Cleaning the previous deployment (if any)...${NC}"
./clean.sh

echo -e "${GREEN}Adjusting some variables.sh parameters...${NC}"
sed -i 's/export USE_TXGEN=[01]/export USE_TXGEN=1/' variables.sh
sed -i 's/export USE_PROXY=[01]/export USE_PROXY=1/' variables.sh
sed -i 's/export USE_ELASTICSEARCH=[01]/export USE_ELASTICSEARCH=1/' variables.sh
sed -i 's/export SOVEREIGN_DEPLOY=[01]/export SOVEREIGN_DEPLOY=1/' variables.sh

echo -e "${GREEN}Generating the configuration files...${NC}"
./config.sh

echo -e "${GREEN}Starting the sovereign chain (in a screen)...${NC}"
screen -L -Logfile sovereignStartLog.txt -d -m -S sovereignStartScreen ./sovereignStart.sh debug

echo -e "${GREEN}Sleeping few minutes so the sovereign chain will begin...${NC}"
sleep 120

echo -e "${GREEN}Starting sending the basic scenario transactions...${NC}"
screen -L -Logfile txgenBasicLog.txt -d -m -S txgenBasicScreen ./sovereign-txgen-basic.sh

echo -e "${GREEN}Starting sending the erc20 scenario transactions...${NC}"
screen -L -Logfile txgenErc20Log.txt -d -m -S txgenErc20Screen ./sovereign-txgen-erc20.sh

echo -e "${GREEN}Finished the sovereign chain deployment. Don't forget to stop it with ./stop.sh at the end.${NC}"

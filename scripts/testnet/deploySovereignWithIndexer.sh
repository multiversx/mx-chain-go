#!/usr/bin/env bash

GREEN='\x1B[0;32m'
NC='\x1B[0m'

echo -e "${GREEN}Making sure any previous deployment is stopped...${NC}"
./stop.sh

echo -e "${GREEN}Cleaning the previous deployment (if any)...${NC}"
./clean.sh

echo -e "${GREEN}Adjusting some variables.sh parameters...${NC}"
sed -i 's/export USE_PROXY=[01]/export USE_PROXY=1/' variables.sh
sed -i 's/export USE_ELASTICSEARCH=[01]/export USE_ELASTICSEARCH=1/' variables.sh
sed -i 's/export SOVEREIGN_DEPLOY=[01]/export SOVEREIGN_DEPLOY=1/' variables.sh

source variables.sh
if [ "$SHARD_VALIDATORCOUNT" -lt 3 ]; then
  sed -i 's/export SHARD_VALIDATORCOUNT=.*/export SHARD_VALIDATORCOUNT=3/' variables.sh
fi

echo -e "${GREEN}Generating the configuration files...${NC}"
./config.sh

echo -e "${GREEN}Starting the sovereign chain (in a screen)...${NC}"
screen -L -Logfile sovereignStartLog.txt -d -m -S sovereignStartScreen ./sovereignStart.sh debug

echo -e "${GREEN}Sleeping few minutes so the sovereign chain will begin...${NC}"
sleep 120

echo -e "${GREEN}Finished the sovereign chain deployment. Don't forget to stop it with ./stop.sh at the end.${NC}"

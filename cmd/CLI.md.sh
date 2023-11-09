#!/usr/bin/env bash

generate() {
    generateForAssessmentTool
    generateForKeyGenerator
    generateForLogViewer
    generateForNode
    generateForSeedNode
    generateForTermUi
}

generateForAssessmentTool() {
    HELP="
# Assessment Tool CLI

The **Assessment Tool** exposes the following Command Line Interface:
$(code)
\$ assessment --help

$(./assessment/assessment --help | head -n -3)
$(code)
"
    echo "$HELP" > ./assessment/CLI.md
}

generateForKeyGenerator() {
    HELP="
# Keygenerator CLI

The **Key generation Tool** exposes the following Command Line Interface:
$(code)
\$ keygenerator --help

$(./keygenerator/keygenerator --help | head -n -3)
$(code)
"
    echo "$HELP" > ./keygenerator/CLI.md
}

generateForLogViewer() {
    HELP="
# Logviewer App

The **MultiversX Logviewer App** exposes the following Command Line Interface:
$(code)
\$ logviewer --help

$(./logviewer/logviewer --help | head -n -3)
$(code)
"
    echo "$HELP" > ./logviewer/CLI.md
}

generateForNode() {
    HELP="
# Node CLI

The **MultiversX Node** exposes the following Command Line Interface:
$(code)
\$ node --help

$(./node/node --help | head -n -3)
$(code)
"
    echo "$HELP" > ./node/CLI.md
}

generateForSeedNode() {
    HELP="
# MultiversX SeedNode CLI

The **MultiversX SeedNode** exposes the following Command Line Interface:
$(code)
\$ seednode --help

$(./seednode/seednode --help | head -n -3)
$(code)
"
    echo "$HELP" > ./seednode/CLI.md
}

generateForTermUi() {
    HELP="
# MultiversX TermUI CLI

The **MultiversX Term UI** exposes the following Command Line Interface:
$(code)
\$ termui --help

$(./termui/termui --help | head -n -3)
$(code)
"
    echo "$HELP" > ./termui/CLI.md
}

code() {
    printf "\n\`\`\`\n"
}

generate

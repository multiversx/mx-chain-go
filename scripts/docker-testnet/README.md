# Setting up a local-testnet with Docker

First and foremost, one needs to build the **seednode** & **node** images. Hence, the **_build.sh_**
script is provided. This can be done, by invoking the script or building the images manually.

```
./build.sh  # (Optional) Can be ignored if you already have the images stored in the local registry.
./start.sh  # Will start the local-testnet.
./clean.sh  # Will stop and remove the containers related to the local-testnet.
```
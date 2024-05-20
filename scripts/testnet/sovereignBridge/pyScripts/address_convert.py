import sys

from multiversx_sdk_core import Address


def main():
    # input arguments
    address = Address.from_bech32(sys.argv[1])

    print(address.to_hex())


if __name__ == "__main__":
    main()

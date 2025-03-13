function recommendGasPrice(gasStationMetadata, senderShard, desiredLevel, transactionDataLength, transactionGasLimit) {
    console.debug(`Recommend gas price for shard = ${senderShard}, level = ${desiredLevel}, data length = ${transactionDataLength}, gas limit = ${transactionGasLimit}`);

    const MIN_GAS_LIMIT = 50_000;
    const MIN_GAS_PRICE = 1_000_000_000;
    const GAS_PER_DATA_BYTE = 1_500;
    const GAS_PRICE_MODIFIER = 0.01;

    const GAS_PRICE_FALLBACK = MIN_GAS_PRICE;
    const MAX_GAS_PRICE = MIN_GAS_PRICE * 30;

    const ppu = gasStationMetadata[senderShard][desiredLevel];
    if (ppu == 0) {
        // No constraint (maybe less transactions in the pool).
        return GAS_PRICE_FALLBACK;
    }

    if (!ppu) {
        console.error(`No target PPU found for shard = ${senderShard}, level = ${desiredLevel}, maybe bad desired level? Using fallback gas price...`);
        return GAS_PRICE_FALLBACK;
    }

    const dataCost = MIN_GAS_LIMIT + transactionDataLength * GAS_PER_DATA_BYTE;
    const executionCost = transactionGasLimit - dataCost;
    const initiallyPaidFee = transactionGasLimit * ppu;
    const gasPrice = Math.floor(initiallyPaidFee / (dataCost + executionCost * GAS_PRICE_MODIFIER));

    if (gasPrice < MIN_GAS_PRICE) {
        console.error(`Gas price ${gasPrice} too low, maybe bad formula? Using min gas price...`);
        return MIN_GAS_PRICE;
    }

    if (gasPrice > MAX_GAS_PRICE) {
        console.warn(`Gas price ${gasPrice} too high, maybe bad formula? Using max gas price...`);
        console.warn(`Estimated fee: ${(dataCost * MAX_GAS_PRICE + executionCost * MAX_GAS_PRICE * GAS_PRICE_MODIFIER) / 1e18} EGLD`)
        return MAX_GAS_PRICE;
    }

    console.debug(`Recommended gas price: ${gasPrice}, which is ${Math.floor(gasPrice / MIN_GAS_PRICE)} times the minimum gas price.`);
    console.debug(`Estimated fee: ${initiallyPaidFee / 1e18} EGLD`)

    return gasPrice;
}

let myGasStationMetadata = [
    {
        "fast": 0,
        "faster": 11760000
    },
    {
        "fast": 1_000_000_000,
        "faster": 1_000_000_000
    },
    {
        "fast": 0,
        "faster": 11760000
    }
];

console.log("=".repeat(50));
console.log("gasStationMetadata", myGasStationMetadata);

console.log("a", "=".repeat(10));
console.log(recommendGasPrice(myGasStationMetadata, 0, "fast", 30, 45_000_000));

console.log("b", "=".repeat(10));
console.log(recommendGasPrice(myGasStationMetadata, 0, "faster", 30, 45_000_000));

console.log("c", "=".repeat(10));
console.log(recommendGasPrice(myGasStationMetadata, 0, "excellentJustLikeMoveBalance", 30, 1_000_000));

console.log("d", "=".repeat(10));
console.log(recommendGasPrice(myGasStationMetadata, 0, "excellentJustLikeMoveBalance", 30, 45_000_000));

console.log("e", "=".repeat(10));
console.log(recommendGasPrice(myGasStationMetadata, 0, "foobar", 5, 45_000_000));

console.log("f", "=".repeat(10));
console.log(recommendGasPrice(myGasStationMetadata, 0, "fast", 100, 600_000_000));

console.log("g", "=".repeat(10));
console.log(recommendGasPrice(myGasStationMetadata, 0, "faster", 600_000_000, 100));

console.log("h (move balance)", "=".repeat(10));
console.log(recommendGasPrice(myGasStationMetadata, 0, "fast", 0, 50_000));

console.log("j (xExchange swap)", "=".repeat(10));
console.log(recommendGasPrice(myGasStationMetadata, 1, "fast", 143, 30_000_000));
console.log(recommendGasPrice(myGasStationMetadata, 1, "faster", 143, 30_000_000));

myGasStationMetadata = [
    {
        "fast": 11760000,
        "faster": 22760000
    },
    {
        "fast": 1_000_000_000,
        "faster": 1_000_000_000
    },
    {
        "fast": 11760000,
        "faster": 22760000
    }
];

console.log("=".repeat(50));
console.log("gasStationMetadata", myGasStationMetadata);

console.log("a", "=".repeat(10));
console.log(recommendGasPrice(myGasStationMetadata, 0, "fast", 30, 45_000_000));

console.log("b", "=".repeat(10));
console.log(recommendGasPrice(myGasStationMetadata, 0, "faster", 30, 45_000_000));

console.log("c", "=".repeat(10));
console.log(recommendGasPrice(myGasStationMetadata, 0, "excellentJustLikeMoveBalance", 30, 1_000_000));

console.log("d", "=".repeat(10));
console.log(recommendGasPrice(myGasStationMetadata, 0, "excellentJustLikeMoveBalance", 30, 45_000_000));

console.log("e", "=".repeat(10));
console.log(recommendGasPrice(myGasStationMetadata, 0, "foobar", 5, 45_000_000));

console.log("f", "=".repeat(10));
console.log(recommendGasPrice(myGasStationMetadata, 0, "fast", 100, 600_000_000));

console.log("g", "=".repeat(10));
console.log(recommendGasPrice(myGasStationMetadata, 0, "faster", 600_000_000, 100));

console.log("h (move balance)", "=".repeat(10));
console.log(recommendGasPrice(myGasStationMetadata, 0, "fast", 0, 50_000));

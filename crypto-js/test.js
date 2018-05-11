

//-------------- keys ---------------------
const ecdsa = new KJUR.crypto.ECDSA({curve: "secp256k1"});
const ecdsaKeyLen = ecdsa.ecparams.keylen / 4;

function PrivateKeyBySecret(secret){
    return normInt(XHash(secret).toString().substring(0, ecdsaKeyLen))
}

function normInt(b) {
    return new BigInteger(b, 16).mod(ecdsa.ecparams.n).add(BigInteger.ONE).toString(16);
}

KJUR.crypto.ECDSA.biRSSigToASN1Sig = function(x, y) {
    return ("000000000000000" + x.toString(16)).slice(-ecdsaKeyLen)
        + ("000000000000000" + y.toString(16)).slice(-ecdsaKeyLen);
};
KJUR.crypto.ECDSA.parseSigHex = function(signHex) {
    return {
        r: new BigInteger(signHex.substr(0, ecdsaKeyLen), 16),
        s: new BigInteger(signHex.substr(ecdsaKeyLen), 16)
    }
};
function PrivateKeyBySecret(secret) {
    const hash = XHash(secret).toString();
    return normInt(hash.substring(0, ecdsaKeyLen)).toString();
}
function PublicKeyByPrivate(prvHex) {
    const m = ecdsa.ecparams.G.multiply(new BigInteger(prvHex, 16));
    return ("000000000000000" + m.getX().toBigInteger().toString(16)).slice(-ecdsaKeyLen)
        + ("000000000000000" + m.getY().toBigInteger().toString(16)).slice(-ecdsaKeyLen);
}


//--------------- TESTING ------------------------------------
// XHash
console.time();
const hash = XHash("abc").toString();
console.timeEnd();

console.log("hash:", hash);
console.assert(hash === "6c23f9a5fc3609c74a37e2fb7982653c97e39f00a8f700f99e3b770bb872bd8b9b819d546a2cf5a2aebc10a28f75886a76ccc8c4f1ec8999652c9bb31ec8c8a7");

// PrivateKeyBySecret
const prv = normInt(hash.substring(0, ecdsaKeyLen)).toString();
const pub = PublicKeyByPrivate(prv);

console.log("prv:", prv);
console.log("pub:", pub);
console.assert(prv === "6c23f9a5fc3609c74a37e2fb7982653c97e39f00a8f700f99e3b770bb872bd8c");
console.assert(pub === "cca27aa571d2838209895c8151cff2ade07e56c0c62a64cd3c84ad73a4287141b207d20a99a5c3169f5f086c3bd05480cad1ad1359a7f01151ed2fd9b2c67601");

//------- alice key
const alicePrv = PrivateKeyBySecret("Alice secret");
console.assert(alicePrv == "8593d5e4dcadd6bb57ed7ff303f49720016daaf41e0cb7d5d688196fd15008d1");

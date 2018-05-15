

// hash (shake512)
const v0=crypto.hash("");
const v1=crypto.hash("abc");

console.log("hash('')      ", v0);
console.log("hash('abc')   ", v1);
console.assert(v0 === "46b9dd2b0ba88d13233b3feb743eeb243fcd52ea62b81b82b50c27646ed5762f");
console.assert(v1 === "483366601360a8771c6863080cc4114d8db44530f8f1e1ee4f94ea37e78b5739");

// xhash
console.time("xhash");
const xhash=crypto.xhash("abc");
console.timeEnd("xhash");
console.log("xhash:", xhash);
console.assert(xhash === "d85f68a0dd6b2ebeb5a60b47b70d2e4b63a842ac0510116e2d52f153e535cd60635f76fd52b34cceb671e0ed11093e923c39ee1a5ff32088ebf5f2415a285eef");

// private key by secret
console.time("prv");
const prv = crypto.privateKeyBySecret("abc");
console.timeEnd("prv");
console.log("prv:", prv);
console.assert(prv === "0xd85f68a0dd6b2ebeb5a60b47b70d2e4b63a842ac0510116e2d52f153e535cd61");

// public key by private
console.time("pub");
const pub = crypto.publicKeyByPrivate(prv);
console.timeEnd("pub");
console.log("pub:", pub);
console.assert(pub === "0x022f86f8c408c20e8bdcef6471676a2157624915355fe662b568ac5e2a2a76fed5d34d4a184176a3e4a28bac7203a860510e363601f7c8f8657067173ed83f6e");

// hex-address by public key
console.time("addr");
const addrHex = crypto.addressByPublic(pub);
console.timeEnd("addr");
console.log("addr:", addrHex);
console.assert(addrHex === "0x9a8a9d2b5766b5c3962f4dd301c01765bdc37a6387f24250");

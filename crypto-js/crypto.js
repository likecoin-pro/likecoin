
const SHA256 = CryptoJS.SHA256;
const SHA512 = CryptoJS.SHA512;
const latin1 = CryptoJS.enc.Latin1;
const hex = CryptoJS.enc.Hex;
const ecdsa = new KJUR.crypto.ECDSA({curve: "secp256k1"});
const ecdsaKeyLen = ecdsa.ecparams.keylen / 4;

const crypto = {

    shake512: function(data) {
        return shake256.create(512).update(data).toString();
    },

    hash: function(data) {
        return shake256.create(256).update(data).toString();
    },

    xhash: function(data) {
        const n = 200003;
        const a = new Array(n);
        for(let i=0; i<n; i++) {
            data = shake256.create(256).update(data).array();
            a[i] = data.slice(-16);
        }
        a.sort(function(a, b){
            for(let i=0;i<64;i++) if(a[i]!==b[i]) return a[i]<b[i]? -1 : 1;
            return 0;
        });
        const h = shake256.create(512);
        for(let i=0;i<n;i++) h.update(a[i]);
        return h.toString();
    },

    privateKeyBySecret: function(secret) {
        return "0x"+normInt(this.xhash(secret).toString().substring(0, ecdsaKeyLen))
    },

    publicKeyByPrivate: function(prv) {
        const m = ecdsa.ecparams.G.multiply(newBigInt(prv));
        return "0x"
            + ("000000000000000" + m.getX().toBigInteger().toString(16)).slice(-ecdsaKeyLen)
            + ("000000000000000" + m.getY().toBigInteger().toString(16)).slice(-ecdsaKeyLen);
    },

    addressByPublic: function(pubHex) {
        let h = hex2arr(pubHex);
        h = shake256.create(512).update(h).array();
        h = shake256.create(512).update(h);
        return "0x"+h.toString().slice(-48);
    },

};

function cutHexPrefix(s) {
    return s.substr(0, 2)==="0x"? s.substr(2) : s;
}

function hex2arr(s) {
    s = cutHexPrefix(s);
    const n = s.length>>1;
    const a = new Array(n);
    for(let i=0; i<n; i++) a[i]=parseInt(s.substr(i<<1, 2), 16);
    return a;
}

function newBigInt(s) {
    return new BigInteger(cutHexPrefix(s), 16);
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


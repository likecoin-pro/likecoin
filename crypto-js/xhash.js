
function XHash(data){
    var n = 84673;
    var a = new Array(n);
    for(var i=0; i<n; i++) {
        data = CryptoJS.SHA256(data);
        a[i] = data.toString(CryptoJS.enc.Latin1);
    }
    a.sort();
    return CryptoJS.SHA512(CryptoJS.enc.Latin1.parse(a.join("")));
}

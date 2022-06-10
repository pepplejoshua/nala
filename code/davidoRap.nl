let davidoRap = fn(iTellYouSayILoveYou, myStuff) {
    let yourStuff = {}
    if (iTellYouSayILoveYou) {
        yourStuff["money"] = myStuff["money"]
        yourStuff["body"] = myStuff["body"]
    }
    yourStuff["account"] = pow(30 * 10, 10)
    return yourStuff
}

puts(davidoRap(true, {"money": 50000000, "body": "body"}))


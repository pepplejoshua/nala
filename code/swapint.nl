let a = 200
let b = 120
let c = 50
let d = 500

let swapInts = fn(first, second) {
    let first = first + second;
    let second = first - second;
    let first = first - second;
    return [first, second]
}

let swapIntsMid = fn(f, s) {
    let m = s
    return [s, f]
}

puts("a before swap: ", a)
puts("b before swap: ", b)
putl()
let res = swapInts(a, b)
let a = res[0]
let b = res[1]
puts("a after swap: ", a)
puts("b after swap: ", b)

putl()
puts("c before swap: ", c)
puts("d before swap: ", d)
putl()
let res = swapIntsMid(c, d)
let c = res[0]
let d = res[1]
puts("c after swap: ", c)
puts("d after swap: ", d)
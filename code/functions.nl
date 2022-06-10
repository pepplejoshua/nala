let map = fn(arr, f) {
    let iter = fn(arr, accum, iter) {
        if (len(arr) == 0) {
            accum
        } else {
            iter(rest(arr), push(accum, f(first(arr))))
        }

    };
    iter(arr, [], iter);
};

let reduce = fn(arr, init, f) {
    let iter = fn(arr, res) {
        if (len(arr) == 0) {
            res
        } else {
            reduce(rest(arr), f(res, first(arr)), f)
        }
    }
    iter(arr, init)
};

let incr = fn(x, in) { return x + in  }

let isInt = fn(x) { type(x) == "INTEGER" }

let isIntArr = fn(arr) { map(arr, isInt) }

let sum = fn(arr) { reduce(arr, 0, incr) };

let product = fn(arr) {
    reduce(arr, 1, fn(in, el) { in * el })
};

let pow = fn(num, times) {
    if (times == 1) {
        return num
    } else {
        return num * pow(num, times - 1)
    }
    
}

let info = {"name": "Nala", "version": "0.0.9", "author": "Iwarilama"};

let fibo = fn(x) { if (x < 2) { return x }; fibo(x - 1) + fibo(x - 2) };

let nums = [1, 2, 3];

let na = fn(a) {
    fn(b) {
        fn(c) {
            a + b + c
        }
    }
}

let xover = "Hello Ellisp from Nala Side!!"

let Nalafullname = fn(f) {
    fn(m) {
        fn(l) {
            "your names are " + f + m + l
        }
    }
}

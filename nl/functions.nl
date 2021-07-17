let map = fn(arr, f) {
    let iter = fn(arr, accum) {
        if (len(arr) == 0) {
            accum
        } else {
            iter(rest(arr), push(accum, f(first(arr))))
        }

    };
    iter(arr, []);
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
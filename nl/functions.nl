let unless = macro(cond, cons, alt) {
				quote(
					if (!unquote(cond)) {
						unquote(cons)
					} else {
						unquote(alt)
					}
				);
			}

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
            iter(rest(arr), f(res, first(arr)))
        }
    }
    iter(arr, init)
};

let sum = fn(arr) {
    reduce(arr, 0, fn(in, el) { in + el })
};

let product = fn(arr) {
    reduce(arr, 1, fn(in, el) { in * el })
};

let info = {"name": "Nala", "version": "0.0.9", "author": "Iwarilama"};

let reverseMinus = macro(a, b) { quote(unquote(b) - unquote(a)); }
let tern = fn(cond, cons, alt) { 
    if (cond) {
        cons
    } else {
        alt
    }
}